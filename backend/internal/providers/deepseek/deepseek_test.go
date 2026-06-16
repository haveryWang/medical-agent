package deepseek

import (
	"context"
	"errors"
	"testing"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/models"
)

func TestConfiguredUsesResolvedDatabaseModelConfig(t *testing.T) {
	client := New(config.Config{}, WithModelConfigResolver(func(context.Context) models.ModelConfig {
		return models.ModelConfig{
			DeepSeekBaseURL:   "https://db.deepseek.example",
			DeepSeekAPIKey:    "db-key",
			DeepSeekChatModel: "deepseek-chat",
		}
	}))

	if !client.Configured(context.Background()) {
		t.Fatal("expected client to be configured from database model config")
	}
}

func TestStreamChatRejectsIncompleteModelConfig(t *testing.T) {
	client := New(config.Config{
		DeepSeekBaseURL:   "https://ark.cn-beijing.volces.com/api/v3",
		DeepSeekChatModel: "deepseek-v4-flash-260425",
	})

	err := client.StreamChat(context.Background(), []Message{{Role: "user", Content: "hello"}}, func(string) error {
		t.Fatal("onDelta should not be called when model config is incomplete")
		return nil
	})
	if !errors.Is(err, ErrModelConfigIncomplete) {
		t.Fatalf("expected ErrModelConfigIncomplete, got %v", err)
	}
}
