package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

var ErrModelConfigIncomplete = errors.New("对话模型配置不完整")

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

func New(cfg config.Config, options ...func(*Client)) *Client {
	client := &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 120 * time.Second},
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
	if strings.TrimSpace(model.DeepSeekBaseURL) != "" {
		cfg.DeepSeekBaseURL = model.DeepSeekBaseURL
	}
	if strings.TrimSpace(model.DeepSeekAPIKey) != "" {
		cfg.DeepSeekAPIKey = model.DeepSeekAPIKey
	}
	if strings.TrimSpace(model.DeepSeekChatModel) != "" {
		cfg.DeepSeekChatModel = model.DeepSeekChatModel
	}
	return cfg
}

func (c *Client) Configured(ctx context.Context) bool {
	cfg := c.resolveConfig(ctx)
	return chatConfigured(cfg)
}

func (c *Client) Health(ctx context.Context) error {
	if !c.Configured(ctx) {
		return chatConfigError()
	}
	return nil
}

func (c *Client) StreamChat(ctx context.Context, messages []Message, onDelta func(string) error) error {
	cfg := c.resolveConfig(ctx)
	if !chatConfigured(cfg) {
		return chatConfigError()
	}
	payload, err := json.Marshal(chatRequest{Model: cfg.DeepSeekChatModel, Messages: messages, Stream: true})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.DeepSeekBaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.DeepSeekAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("DeepSeek 返回状态码 %d: %s", res.StatusCode, string(body))
	}
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return nil
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				if err := onDelta(choice.Delta.Content); err != nil {
					return err
				}
			}
		}
	}
	return scanner.Err()
}

func chatConfigured(cfg config.Config) bool {
	return strings.TrimSpace(cfg.DeepSeekBaseURL) != "" &&
		strings.TrimSpace(cfg.DeepSeekAPIKey) != "" &&
		strings.TrimSpace(cfg.DeepSeekChatModel) != ""
}

func chatConfigError() error {
	return fmt.Errorf("%w，请在系统设置中配置对话 Base URL、API Key 和模型", ErrModelConfigIncomplete)
}

func (c *Client) localStream(ctx context.Context, messages []Message, onDelta func(string) error) error {
	question := ""
	contextText := ""
	for _, msg := range messages {
		if msg.Role == "user" {
			question = msg.Content
		}
		if msg.Role == "system" {
			contextText = msg.Content
		}
	}
	answer := "当前未配置 DeepSeek API Key，以下为本地演示回答。"
	if contextText != "" {
		answer += "我已根据知识库检索内容组织回复："
	}
	if question != "" {
		answer += "关于「" + question + "」，建议结合引用来源进行核对，并由医生制定个体化方案。"
	}
	parts := []rune(answer)
	for i := 0; i < len(parts); i += 8 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		end := i + 8
		if end > len(parts) {
			end = len(parts)
		}
		if err := onDelta(string(parts[i:end])); err != nil {
			return err
		}
		time.Sleep(40 * time.Millisecond)
	}
	return nil
}
