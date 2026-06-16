package qwen

import (
	"context"
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
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"embedding":[0.1,0.2,0.3]}]}`))
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

	_, err := client.Embed(context.Background(), []string{"hello"})
	if err != nil {
		t.Fatalf("Embed returned error: %v", err)
	}
	if gotPath != "/embeddings/multimodal" {
		t.Fatalf("request path = %q, want /embeddings/multimodal", gotPath)
	}
	if !strings.Contains(gotBody, `"type":"text"`) || !strings.Contains(gotBody, `"text":"hello"`) {
		t.Fatalf("request body = %s, want Volcengine multimodal text input", gotBody)
	}
}
