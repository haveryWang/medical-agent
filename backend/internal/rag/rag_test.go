package rag

import (
	"context"
	"testing"

	"medical-agent/backend/internal/config"
)

func TestRetrieveSkipsEnhancementWhenNoKnowledgeBaseSelected(t *testing.T) {
	service := New(config.Config{}, nil, nil, nil, nil)

	retrieval, err := service.Retrieve(context.Background(), "hello", nil)
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}
	if retrieval.Context != "" {
		t.Fatalf("expected empty context when retrieval is disabled, got %q", retrieval.Context)
	}
	if len(retrieval.Citations) != 0 {
		t.Fatalf("expected no citations when retrieval is disabled, got %#v", retrieval.Citations)
	}
}
