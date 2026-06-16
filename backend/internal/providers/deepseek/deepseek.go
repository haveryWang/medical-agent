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
)

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

func New(cfg config.Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) Configured() bool {
	return c.cfg.DeepSeekAPIKey != "" && c.cfg.DeepSeekBaseURL != "" && c.cfg.DeepSeekChatModel != ""
}

func (c *Client) Health(ctx context.Context) error {
	if !c.Configured() {
		return errors.New("DeepSeek 未配置，需设置 DEEPSEEK_API_KEY")
	}
	return nil
}

func (c *Client) StreamChat(ctx context.Context, messages []Message, onDelta func(string) error) error {
	if !c.Configured() {
		return c.localStream(ctx, messages, onDelta)
	}
	payload, err := json.Marshal(chatRequest{Model: c.cfg.DeepSeekChatModel, Messages: messages, Stream: true})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.DeepSeekBaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.DeepSeekAPIKey)
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
