package qwen

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

var ErrModelConfigIncomplete = errors.New("向量模型配置不完整")

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

type multimodalEmbeddingResponse struct {
	Data struct {
		Embedding embeddingVector `json:"embedding"`
	} `json:"data"`
}

type embeddingVector []float64

func (v *embeddingVector) UnmarshalJSON(data []byte) error {
	var flat []float64
	if err := json.Unmarshal(data, &flat); err == nil {
		*v = flat
		return nil
	}
	var nested [][]float64
	if err := json.Unmarshal(data, &nested); err != nil {
		return err
	}
	if len(nested) == 0 {
		*v = nil
		return nil
	}
	*v = nested[0]
	return nil
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
	return embeddingConfigured(cfg)
}

func (c *Client) Health(ctx context.Context) error {
	if !c.Configured(ctx) {
		return embeddingConfigError()
	}
	_, err := c.Embed(ctx, []string{"健康检查"})
	return err
}

func (c *Client) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	cfg := c.resolveConfig(ctx)
	if !embeddingConfigured(cfg) {
		return nil, embeddingConfigError()
	}
	return c.remoteEmbed(ctx, cfg, texts)
}

func embeddingConfigured(cfg config.Config) bool {
	return strings.TrimSpace(cfg.QwenEmbeddingBaseURL) != "" &&
		strings.TrimSpace(cfg.QwenEmbeddingAPIKey) != "" &&
		strings.TrimSpace(cfg.QwenEmbeddingModel) != "" &&
		cfg.QwenEmbeddingDimension > 0
}

func embeddingConfigError() error {
	return fmt.Errorf("%w，请在系统设置中配置向量 Base URL、API Key、模型和维度", ErrModelConfigIncomplete)
}

func (c *Client) remoteEmbed(ctx context.Context, cfg config.Config, texts []string) ([][]float64, error) {
	if isVolcengineMultimodalEmbeddingModel(cfg.QwenEmbeddingModel) {
		return c.remoteMultimodalEmbed(ctx, cfg, texts)
	}
	payload, path, err := embeddingPayload(cfg.QwenEmbeddingModel, texts)
	if err != nil {
		return nil, err
	}
	url := strings.TrimRight(cfg.QwenEmbeddingBaseURL, "/") + path
	body, err := c.postEmbedding(ctx, url, cfg.QwenEmbeddingAPIKey, payload)
	if err != nil {
		return nil, err
	}
	var decoded embeddingResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
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

func (c *Client) remoteMultimodalEmbed(ctx context.Context, cfg config.Config, texts []string) ([][]float64, error) {
	url := strings.TrimRight(cfg.QwenEmbeddingBaseURL, "/") + "/embeddings/multimodal"
	vectors := make([][]float64, 0, len(texts))
	for _, text := range texts {
		payload, err := json.Marshal(multimodalEmbeddingRequest{
			Model: cfg.QwenEmbeddingModel,
			Input: []multimodalEmbeddingItem{{Type: "text", Text: text}},
		})
		if err != nil {
			return nil, err
		}
		body, err := c.postEmbedding(ctx, url, cfg.QwenEmbeddingAPIKey, payload)
		if err != nil {
			return nil, err
		}
		var decoded multimodalEmbeddingResponse
		if err := json.Unmarshal(body, &decoded); err != nil {
			return nil, err
		}
		vector := []float64(decoded.Data.Embedding)
		if len(vector) != cfg.QwenEmbeddingDimension {
			return nil, fmt.Errorf("向量模型维度不匹配: got %d want %d", len(vector), cfg.QwenEmbeddingDimension)
		}
		vectors = append(vectors, vector)
	}
	return vectors, nil
}

func (c *Client) postEmbedding(ctx context.Context, url string, apiKey string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, modelHTTPError(res.StatusCode, body)
	}
	return body, nil
}

func embeddingPayload(model string, texts []string) ([]byte, string, error) {
	payload, err := json.Marshal(embeddingRequest{Model: model, Input: texts})
	return payload, "/embeddings", err
}

func isVolcengineMultimodalEmbeddingModel(model string) bool {
	return model == "doubao-embedding-vision-251215"
}

func modelHTTPError(statusCode int, body []byte) error {
	message := strings.TrimSpace(string(body))
	if message == "" {
		return fmt.Errorf("向量模型返回状态码 %d", statusCode)
	}
	if len([]rune(message)) > 300 {
		runes := []rune(message)
		message = string(runes[:300]) + "..."
	}
	return fmt.Errorf("向量模型返回状态码 %d: %s", statusCode, message)
}

func deterministicVector(text string, dimension int) []float64 {
	if dimension <= 0 {
		dimension = config.DefaultQwenEmbeddingDimension
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
