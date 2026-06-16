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
	"medical-agent/backend/internal/models"
)

type Client struct {
	cfg           config.Config
	httpClient    *http.Client
	modelResolver func(context.Context) models.ModelConfig
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type multimodalEmbeddingRequest struct {
	Model string                    `json:"model"`
	Input []multimodalEmbeddingItem `json:"input"`
}

type multimodalEmbeddingItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func New(cfg config.Config, options ...func(*Client)) *Client {
	client := &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
	for _, option := range options {
		option(client)
	}
	return client
}

func WithModelConfigResolver(resolver func(context.Context) models.ModelConfig) func(*Client) {
	return func(client *Client) {
		client.modelResolver = resolver
	}
}

func (c *Client) resolveConfig(ctx context.Context) config.Config {
	cfg := c.cfg
	if c.modelResolver == nil {
		return cfg
	}
	model := c.modelResolver(ctx)
	if strings.TrimSpace(model.QwenEmbeddingBaseURL) != "" {
		cfg.QwenEmbeddingBaseURL = model.QwenEmbeddingBaseURL
	}
	if strings.TrimSpace(model.QwenEmbeddingAPIKey) != "" {
		cfg.QwenEmbeddingAPIKey = model.QwenEmbeddingAPIKey
	}
	if strings.TrimSpace(model.QwenEmbeddingModel) != "" {
		cfg.QwenEmbeddingModel = model.QwenEmbeddingModel
	}
	if model.QwenEmbeddingDimension > 0 {
		cfg.QwenEmbeddingDimension = model.QwenEmbeddingDimension
	}
	return cfg
}

func (c *Client) Configured(ctx context.Context) bool {
	cfg := c.resolveConfig(ctx)
	return cfg.QwenEmbeddingBaseURL != "" && cfg.QwenEmbeddingAPIKey != "" && cfg.QwenEmbeddingModel != "" && cfg.QwenEmbeddingDimension > 0
}

func (c *Client) Health(ctx context.Context) error {
	if !c.Configured(ctx) {
		return errors.New("向量模型未配置，需设置 VOLCENGINE_API_KEY 或 QWEN_EMBEDDING_BASE_URL/QWEN_EMBEDDING_API_KEY/QWEN_EMBEDDING_MODEL/QWEN_EMBEDDING_DIMENSION")
	}
	_, err := c.Embed(ctx, []string{"健康检查"})
	return err
}

func (c *Client) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	cfg := c.resolveConfig(ctx)
	if cfg.QwenEmbeddingBaseURL != "" && cfg.QwenEmbeddingAPIKey != "" && cfg.QwenEmbeddingModel != "" && cfg.QwenEmbeddingDimension > 0 {
		vectors, err := c.remoteEmbed(ctx, cfg, texts)
		if err == nil {
			return vectors, nil
		}
		return nil, err
	}
	vectors := make([][]float64, 0, len(texts))
	for _, text := range texts {
		vectors = append(vectors, deterministicVector(text, cfg.QwenEmbeddingDimension))
	}
	return vectors, nil
}

func (c *Client) remoteEmbed(ctx context.Context, cfg config.Config, texts []string) ([][]float64, error) {
	payload, path, err := embeddingPayload(cfg.QwenEmbeddingModel, texts)
	if err != nil {
		return nil, err
	}
	url := strings.TrimRight(cfg.QwenEmbeddingBaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.QwenEmbeddingAPIKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("向量模型返回状态码 %d", res.StatusCode)
	}
	var decoded embeddingResponse
	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if len(decoded.Data) != len(texts) {
		return nil, fmt.Errorf("向量模型返回数量不匹配: got %d want %d", len(decoded.Data), len(texts))
	}
	vectors := make([][]float64, 0, len(decoded.Data))
	for _, item := range decoded.Data {
		if len(item.Embedding) != cfg.QwenEmbeddingDimension {
			return nil, fmt.Errorf("向量模型维度不匹配: got %d want %d", len(item.Embedding), cfg.QwenEmbeddingDimension)
		}
		vectors = append(vectors, item.Embedding)
	}
	return vectors, nil
}

func embeddingPayload(model string, texts []string) ([]byte, string, error) {
	if model == "doubao-embedding-vision-251215" {
		input := make([]multimodalEmbeddingItem, 0, len(texts))
		for _, text := range texts {
			input = append(input, multimodalEmbeddingItem{Type: "text", Text: text})
		}
		payload, err := json.Marshal(multimodalEmbeddingRequest{Model: model, Input: input})
		return payload, "/embeddings/multimodal", err
	}
	payload, err := json.Marshal(embeddingRequest{Model: model, Input: texts})
	return payload, "/embeddings", err
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
