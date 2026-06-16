package qwen

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"medical-agent/backend/internal/config"
)

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func New(cfg config.Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Configured() bool {
	return c.cfg.QwenEmbeddingBaseURL != "" && c.cfg.QwenEmbeddingAPIKey != "" && c.cfg.QwenEmbeddingModel != "" && c.cfg.QwenEmbeddingDimension > 0
}

func (c *Client) Health(ctx context.Context) error {
	if !c.Configured() {
		return errors.New("Qwen3-Embedding 未配置，需设置 QWEN_EMBEDDING_BASE_URL/QWEN_EMBEDDING_API_KEY/QWEN_EMBEDDING_MODEL/QWEN_EMBEDDING_DIMENSION")
	}
	_, err := c.Embed(ctx, []string{"健康检查"})
	return err
}

func (c *Client) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	if c.Configured() {
		vectors, err := c.remoteEmbed(ctx, texts)
		if err == nil {
			return vectors, nil
		}
		return nil, err
	}
	vectors := make([][]float64, 0, len(texts))
	for _, text := range texts {
		vectors = append(vectors, deterministicVector(text, c.cfg.QwenEmbeddingDimension))
	}
	return vectors, nil
}

func (c *Client) remoteEmbed(ctx context.Context, texts []string) ([][]float64, error) {
	payload, err := json.Marshal(embeddingRequest{Model: c.cfg.QwenEmbeddingModel, Input: texts})
	if err != nil {
		return nil, err
	}
	url := strings.TrimRight(c.cfg.QwenEmbeddingBaseURL, "/") + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.QwenEmbeddingAPIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("Qwen3-Embedding 返回状态码 %d", res.StatusCode)
	}
	var decoded embeddingResponse
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if len(decoded.Data) != len(texts) {
		return nil, fmt.Errorf("Qwen3-Embedding 返回数量不匹配: got %d want %d", len(decoded.Data), len(texts))
	}
	vectors := make([][]float64, 0, len(decoded.Data))
	for _, item := range decoded.Data {
		if len(item.Embedding) != c.cfg.QwenEmbeddingDimension {
			return nil, fmt.Errorf("Qwen3-Embedding 向量维度不匹配: got %d want %d", len(item.Embedding), c.cfg.QwenEmbeddingDimension)
		}
		vectors = append(vectors, item.Embedding)
	}
	return vectors, nil
}

func deterministicVector(text string, dimension int) []float64 {
	if dimension <= 0 {
		dimension = 1024
	}
	vector := make([]float64, dimension)
	for i := 0; i < dimension; i++ {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", text, i)))
		raw := binary.BigEndian.Uint64(sum[:8])
		vector[i] = (float64(raw%20000) - 10000) / 10000
	}
	var norm float64
	for _, v := range vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return vector
	}
	for i := range vector {
		vector[i] = vector[i] / norm
	}
	return vector
}
