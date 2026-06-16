package store

import (
	"context"
	"testing"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

type fakeModelConfigStore struct {
	cfg models.ModelConfig
}

func (s *fakeModelConfigStore) GetModelConfig(ctx context.Context, fallback config.Config) (models.ModelConfig, error) {
	return mergeModelConfig(s.cfg, fallback), nil
}

func TestMergeModelConfigPrefersDatabaseValuesIncludingSecrets(t *testing.T) {
	fallback := config.Config{
		DeepSeekBaseURL:        "https://env.deepseek.example",
		DeepSeekAPIKey:         "env-deepseek-key",
		DeepSeekChatModel:      "env-chat-model",
		QwenEmbeddingBaseURL:   "https://env.qwen.example",
		QwenEmbeddingAPIKey:    "env-qwen-key",
		QwenEmbeddingModel:     "env-embedding-model",
		QwenEmbeddingDimension: 1024,
	}
	db := models.ModelConfig{
		DeepSeekBaseURL:        "https://db.deepseek.example",
		DeepSeekAPIKey:         "db-deepseek-key",
		DeepSeekChatModel:      "db-chat-model",
		QwenEmbeddingBaseURL:   "https://db.qwen.example",
		QwenEmbeddingAPIKey:    "db-qwen-key",
		QwenEmbeddingModel:     "db-embedding-model",
		QwenEmbeddingDimension: 1536,
	}

	got := mergeModelConfig(db, fallback)

	if got.DeepSeekBaseURL != db.DeepSeekBaseURL {
		t.Fatalf("DeepSeek base URL = %q, want %q", got.DeepSeekBaseURL, db.DeepSeekBaseURL)
	}
	if got.DeepSeekAPIKey != db.DeepSeekAPIKey {
		t.Fatalf("DeepSeek API key = %q, want database key", got.DeepSeekAPIKey)
	}
	if got.DeepSeekChatModel != db.DeepSeekChatModel {
		t.Fatalf("DeepSeek chat model = %q, want %q", got.DeepSeekChatModel, db.DeepSeekChatModel)
	}
	if got.QwenEmbeddingBaseURL != db.QwenEmbeddingBaseURL {
		t.Fatalf("Qwen base URL = %q, want %q", got.QwenEmbeddingBaseURL, db.QwenEmbeddingBaseURL)
	}
	if got.QwenEmbeddingAPIKey != db.QwenEmbeddingAPIKey {
		t.Fatalf("Qwen API key = %q, want database key", got.QwenEmbeddingAPIKey)
	}
	if got.QwenEmbeddingModel != db.QwenEmbeddingModel {
		t.Fatalf("Qwen model = %q, want %q", got.QwenEmbeddingModel, db.QwenEmbeddingModel)
	}
	if got.QwenEmbeddingDimension != db.QwenEmbeddingDimension {
		t.Fatalf("Qwen dimension = %d, want %d", got.QwenEmbeddingDimension, db.QwenEmbeddingDimension)
	}
}

func TestMergeModelConfigFallsBackToEnvironmentForBlankFields(t *testing.T) {
	fallback := config.Config{
		DeepSeekBaseURL:        "https://env.deepseek.example",
		DeepSeekAPIKey:         "env-deepseek-key",
		DeepSeekChatModel:      "env-chat-model",
		QwenEmbeddingBaseURL:   "https://env.qwen.example",
		QwenEmbeddingAPIKey:    "env-qwen-key",
		QwenEmbeddingModel:     "env-embedding-model",
		QwenEmbeddingDimension: 1024,
	}

	got := mergeModelConfig(models.ModelConfig{}, fallback)

	if got.DeepSeekBaseURL != fallback.DeepSeekBaseURL {
		t.Fatalf("DeepSeek base URL = %q, want fallback", got.DeepSeekBaseURL)
	}
	if got.DeepSeekAPIKey != fallback.DeepSeekAPIKey {
		t.Fatalf("DeepSeek API key = %q, want fallback", got.DeepSeekAPIKey)
	}
	if got.DeepSeekChatModel != fallback.DeepSeekChatModel {
		t.Fatalf("DeepSeek model = %q, want fallback", got.DeepSeekChatModel)
	}
	if got.QwenEmbeddingBaseURL != fallback.QwenEmbeddingBaseURL {
		t.Fatalf("Qwen base URL = %q, want fallback", got.QwenEmbeddingBaseURL)
	}
	if got.QwenEmbeddingAPIKey != fallback.QwenEmbeddingAPIKey {
		t.Fatalf("Qwen API key = %q, want fallback", got.QwenEmbeddingAPIKey)
	}
	if got.QwenEmbeddingModel != fallback.QwenEmbeddingModel {
		t.Fatalf("Qwen model = %q, want fallback", got.QwenEmbeddingModel)
	}
	if got.QwenEmbeddingDimension != fallback.QwenEmbeddingDimension {
		t.Fatalf("Qwen dimension = %d, want fallback", got.QwenEmbeddingDimension)
	}
}

func TestDefaultModelConfigUsesVolcengineProvider(t *testing.T) {
	fallback := config.Config{
		DeepSeekBaseURL:        "https://ark.cn-beijing.volces.com/api/v3",
		DeepSeekAPIKey:         "volcengine-key",
		DeepSeekChatModel:      "DeepSeek-V4-flash",
		QwenEmbeddingBaseURL:   "https://ark.cn-beijing.volces.com/api/v3",
		QwenEmbeddingAPIKey:    "volcengine-key",
		QwenEmbeddingModel:     "doubao-embedding-vision-251215",
		QwenEmbeddingDimension: 2048,
	}

	got := defaultModelConfig(fallback)

	if got.DeepSeekBaseURL != "https://ark.cn-beijing.volces.com/api/v3" {
		t.Fatalf("DeepSeekBaseURL = %q, want Volcengine Ark base URL", got.DeepSeekBaseURL)
	}
	if got.DeepSeekChatModel != "DeepSeek-V4-flash" {
		t.Fatalf("DeepSeekChatModel = %q, want DeepSeek-V4-flash", got.DeepSeekChatModel)
	}
	if got.QwenEmbeddingBaseURL != "https://ark.cn-beijing.volces.com/api/v3" {
		t.Fatalf("QwenEmbeddingBaseURL = %q, want Volcengine Ark base URL", got.QwenEmbeddingBaseURL)
	}
	if got.QwenEmbeddingModel != "doubao-embedding-vision-251215" {
		t.Fatalf("QwenEmbeddingModel = %q, want doubao-embedding-vision-251215", got.QwenEmbeddingModel)
	}
	if got.QwenEmbeddingDimension != 2048 {
		t.Fatalf("QwenEmbeddingDimension = %d, want 2048", got.QwenEmbeddingDimension)
	}
}

func TestMergeModelConfigCorrectsDoubaoVisionDimension(t *testing.T) {
	fallback := config.Config{
		QwenEmbeddingBaseURL:   "https://ark.cn-beijing.volces.com/api/v3",
		QwenEmbeddingAPIKey:    "volcengine-key",
		QwenEmbeddingModel:     "doubao-embedding-vision-251215",
		QwenEmbeddingDimension: 2048,
	}
	got := mergeModelConfig(models.ModelConfig{
		QwenEmbeddingBaseURL:   "https://ark.cn-beijing.volces.com/api/v3",
		QwenEmbeddingAPIKey:    "db-key",
		QwenEmbeddingModel:     "doubao-embedding-vision-251215",
		QwenEmbeddingDimension: 1024,
	}, fallback)

	if got.QwenEmbeddingDimension != 2048 {
		t.Fatalf("QwenEmbeddingDimension = %d, want corrected 2048", got.QwenEmbeddingDimension)
	}
}

func TestLegacyDefaultModelConfigFilterTargetsOldProviders(t *testing.T) {
	filter := legacyDefaultModelConfigFilter()
	encoded := filter["$or"]
	if encoded == nil {
		t.Fatal("legacy filter should contain $or clauses")
	}
	clauses, ok := encoded.([]bson.M)
	if !ok || len(clauses) == 0 {
		t.Fatalf("legacy filter clauses = %#v, want non-empty []bson.M", encoded)
	}
	var foundOldDeepSeek bool
	var foundOldQwen bool
	for _, clause := range clauses {
		if _, ok := clause["deepSeekBaseUrl"]; ok {
			foundOldDeepSeek = true
		}
		if _, ok := clause["qwenEmbeddingModel"]; ok {
			foundOldQwen = true
		}
	}
	if !foundOldDeepSeek || !foundOldQwen {
		t.Fatalf("legacy filter = %#v, want DeepSeek and Qwen legacy clauses", filter)
	}
}
