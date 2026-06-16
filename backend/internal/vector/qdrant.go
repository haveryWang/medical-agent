package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"medical-agent/backend/internal/config"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

type Point struct {
	ID      string         `json:"id"`
	Vector  []float64      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

type SearchResult struct {
	ID    string
	Score float64
}

func New(cfg config.Config) *Client {
	return &Client{cfg: cfg, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func PointIDFromObjectID(id primitive.ObjectID) string {
	hex := id.Hex()
	return fmt.Sprintf("%s-%s-%s-%s-%s00000000", hex[0:8], hex[8:12], hex[12:16], hex[16:20], hex[20:24])
}

func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.QdrantURL+"/collections", nil)
	if err != nil {
		return err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Qdrant 返回状态码 %d", res.StatusCode)
	}
	return nil
}

func (c *Client) EnsureCollection(ctx context.Context) error {
	body := map[string]any{
		"vectors": map[string]any{
			"size":     c.cfg.QwenEmbeddingDimension,
			"distance": "Cosine",
		},
	}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.collectionURL(), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusConflict {
		return nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Qdrant collection 初始化失败: %d", res.StatusCode)
	}
	return nil
}

func (c *Client) Upsert(ctx context.Context, points []Point) error {
	if len(points) == 0 {
		return nil
	}
	body := map[string]any{"points": points}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.collectionURL()+"/points?wait=true", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Qdrant upsert 失败: %d", res.StatusCode)
	}
	return nil
}

func (c *Client) DeletePoints(ctx context.Context, pointIDs []string) error {
	if len(pointIDs) == 0 {
		return nil
	}
	body := map[string]any{"points": pointIDs}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.collectionURL()+"/points/delete?wait=true", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Qdrant 删除向量失败: %d", res.StatusCode)
	}
	return nil
}

func (c *Client) Search(ctx context.Context, vector []float64, knowledgeBaseIDs []string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 5
	}
	filter := map[string]any{}
	if len(knowledgeBaseIDs) > 0 {
		filter = map[string]any{
			"must": []map[string]any{{
				"key":   "knowledgeBaseId",
				"match": map[string]any{"any": knowledgeBaseIDs},
			}},
		}
	}
	body := map[string]any{"vector": vector, "limit": limit, "with_payload": true}
	if len(filter) > 0 {
		body["filter"] = filter
	}
	payload, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.collectionURL()+"/points/search", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("Qdrant search 失败: %d", res.StatusCode)
	}
	var decoded struct {
		Result []struct {
			ID    any     `json:"id"`
			Score float64 `json:"score"`
		} `json:"result"`
	}
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0, len(decoded.Result))
	for _, item := range decoded.Result {
		results = append(results, SearchResult{ID: fmt.Sprint(item.ID), Score: item.Score})
	}
	return results, nil
}

func (c *Client) collectionURL() string {
	return c.cfg.QdrantURL + "/collections/" + c.cfg.QdrantCollection
}
