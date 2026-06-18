package rag

import (
	"context"
	"strings"
	"testing"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
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

func TestBuildPromptContextIncludesAllModelMessages(t *testing.T) {
	messages := BuildPromptMessages("如何执行政策？", Retrieval{Context: "[1] 政策文件\n原文片段"})
	context := FormatPromptContext(messages)

	for _, want := range []string{
		"Role: system",
		"你是医院行政智策平台的智能问答助手。",
		"知识库上下文：",
		"[1] 政策文件\n原文片段",
		"Role: user",
		"如何执行政策？",
	} {
		if !strings.Contains(context, want) {
			t.Fatalf("prompt context missing %q in:\n%s", want, context)
		}
	}
}

func TestRetrievalScopesUseKnowledgeBaseTopK(t *testing.T) {
	firstID := primitive.NewObjectID()
	secondID := primitive.NewObjectID()
	scopes := retrievalScopesFromKnowledgeBases(
		[]primitive.ObjectID{firstID, secondID},
		map[primitive.ObjectID]models.KnowledgeBase{
			firstID:  {RetrievalTopK: 8},
			secondID: {RetrievalTopK: 12},
		},
		5,
	)

	if len(scopes) != 2 {
		t.Fatalf("expected 2 retrieval scopes, got %d", len(scopes))
	}
	if scopes[0].TopK != 8 {
		t.Fatalf("expected first KB topK 8, got %d", scopes[0].TopK)
	}
	if scopes[1].TopK != 12 {
		t.Fatalf("expected second KB topK 12, got %d", scopes[1].TopK)
	}
}

func TestRetrievalScopesFallbackToConfigTopK(t *testing.T) {
	firstID := primitive.NewObjectID()
	secondID := primitive.NewObjectID()
	scopes := retrievalScopesFromKnowledgeBases(
		[]primitive.ObjectID{firstID, secondID},
		map[primitive.ObjectID]models.KnowledgeBase{
			firstID: {RetrievalTopK: 0},
		},
		7,
	)

	if len(scopes) != 2 {
		t.Fatalf("expected 2 retrieval scopes, got %d", len(scopes))
	}
	for _, scope := range scopes {
		if scope.TopK != 7 {
			t.Fatalf("expected fallback topK 7, got %d", scope.TopK)
		}
	}
}
