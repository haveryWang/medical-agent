package vector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"medical-agent/backend/internal/config"
)

func TestEnsureCollectionTreatsAlreadyExistsAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	}))
	defer server.Close()

	client := New(config.Config{
		QdrantURL:              server.URL,
		QdrantCollection:       "medical_agent_chunks",
		QwenEmbeddingDimension: 1024,
	})

	if err := client.EnsureCollection(context.Background()); err != nil {
		t.Fatalf("EnsureCollection returned error for existing collection: %v", err)
	}
}

func TestDeletePointsSendsPointIDsToQdrant(t *testing.T) {
	var received struct {
		Points []string `json:"points"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/collections/medical_agent_chunks/points/delete" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("wait") != "true" {
			t.Fatalf("expected wait=true, got %q", r.URL.RawQuery)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(config.Config{
		QdrantURL:        server.URL,
		QdrantCollection: "medical_agent_chunks",
	})

	if err := client.DeletePoints(context.Background(), []string{"point-a", "point-b"}); err != nil {
		t.Fatalf("DeletePoints returned error: %v", err)
	}
	if len(received.Points) != 2 || received.Points[0] != "point-a" || received.Points[1] != "point-b" {
		t.Fatalf("received points = %#v", received.Points)
	}
}
