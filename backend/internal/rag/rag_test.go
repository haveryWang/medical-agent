package rag

import (
	"context"
	"strings"
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
