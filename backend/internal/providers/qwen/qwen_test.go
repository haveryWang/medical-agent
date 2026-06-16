package qwen

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/models"
)

func TestEmbedUsesResolvedDatabaseModelConfig(t *testing.T) {
	var gotPath string
	var gotAuth string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"embedding":[0.1,0.2,0.3]}]}`))
	}))
	defer server.Close()

	client := New(config.Config{}, WithModelConfigResolver(func(context.Context) models.ModelConfig {
		return models.ModelConfig{
			QwenEmbeddingBaseURL:   server.URL,
			QwenEmbeddingAPIKey:    "db-qwen-key",
			QwenEmbeddingModel:     "db-embedding-model",
			QwenEmbeddingDimension: 3,
		}
	}))

	vectors, err := client.Embed(context.Background(), []string{"hello"})
	if err != nil {
		t.Fatalf("Embed returned error: %v", err)
	}
	if len(vectors) != 1 || len(vectors[0]) != 3 {
		t.Fatalf("vectors = %#v, want one 3-dimensional vector", vectors)
	}
	if gotAuth != "Bearer db-qwen-key" {
		t.Fatalf("Authorization = %q, want database key", gotAuth)
	}
	if gotPath != "/embeddings" {
		t.Fatalf("request path = %q, want /embeddings", gotPath)
	}
	if !strings.Contains(gotBody, `"model":"db-embedding-model"`) {
		t.Fatalf("request body = %s, want database model", gotBody)
	}
}

func TestEmbedUsesVolcengineMultimodalRequestForDoubaoVision(t *testing.T) {
	var gotPath string
	var gotBodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBodies = append(gotBodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"embedding":[0.1,0.2,0.3]}}`))
	}))
	defer server.Close()

	client := New(config.Config{}, WithModelConfigResolver(func(context.Context) models.ModelConfig {
		return models.ModelConfig{
			QwenEmbeddingBaseURL:   server.URL,
			QwenEmbeddingAPIKey:    "volcengine-key",
			QwenEmbeddingModel:     "doubao-embedding-vision-251215",
			QwenEmbeddingDimension: 3,
		}
	}))

	vectors, err := client.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatalf("Embed returned error: %v", err)
	}
	if len(vectors) != 2 {
		t.Fatalf("vectors length = %d, want 2", len(vectors))
	}
	if gotPath != "/embeddings/multimodal" {
		t.Fatalf("request path = %q, want /embeddings/multimodal", gotPath)
	}
	if len(gotBodies) != 2 {
		t.Fatalf("request count = %d, want one request per text", len(gotBodies))
	}
	if !strings.Contains(gotBodies[0], `"type":"text"`) || !strings.Contains(gotBodies[0], `"text":"hello"`) {
		t.Fatalf("request body = %s, want first Volcengine multimodal text input", gotBodies[0])
	}
	if !strings.Contains(gotBodies[1], `"type":"text"`) || !strings.Contains(gotBodies[1], `"text":"world"`) {
		t.Fatalf("request body = %s, want second Volcengine multimodal text input", gotBodies[1])
	}
}

func TestEmbedRejectsIncompleteModelConfig(t *testing.T) {
	client := New(config.Config{
		QwenEmbeddingBaseURL:   "https://ark.cn-beijing.volces.com/api/v3",
		QwenEmbeddingModel:     "doubao-embedding-vision-251215",
		QwenEmbeddingDimension: 2048,
	})

	_, err := client.Embed(context.Background(), []string{"hello"})
	if !errors.Is(err, ErrModelConfigIncomplete) {
		t.Fatalf("expected ErrModelConfigIncomplete, got %v", err)
	}
}
