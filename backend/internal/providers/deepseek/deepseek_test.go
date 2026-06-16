package deepseek

import (
	"context"
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
