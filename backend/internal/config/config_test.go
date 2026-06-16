package config

import "testing"

func TestLoadDefaultsToVolcengineModels(t *testing.T) {
	t.Setenv("DEEPSEEK_BASE_URL", "")
	t.Setenv("DEEPSEEK_CHAT_MODEL", "")
	t.Setenv("QWEN_EMBEDDING_BASE_URL", "")
	t.Setenv("QWEN_EMBEDDING_MODEL", "")
	t.Setenv("QWEN_EMBEDDING_DIMENSION", "")
	t.Setenv("VOLCENGINE_API_KEY", "")

	cfg := Load()

	if cfg.DeepSeekBaseURL != "https://ark.cn-beijing.volces.com/api/v3" {
		t.Fatalf("DeepSeekBaseURL = %q, want Volcengine Ark base URL", cfg.DeepSeekBaseURL)
	}
	if cfg.DeepSeekChatModel != "DeepSeek-V4-flash" {
		t.Fatalf("DeepSeekChatModel = %q, want DeepSeek-V4-flash", cfg.DeepSeekChatModel)
	}
	if cfg.QwenEmbeddingBaseURL != "https://ark.cn-beijing.volces.com/api/v3" {
		t.Fatalf("QwenEmbeddingBaseURL = %q, want Volcengine Ark base URL", cfg.QwenEmbeddingBaseURL)
	}
	if cfg.QwenEmbeddingModel != "doubao-embedding-vision-251215" {
		t.Fatalf("QwenEmbeddingModel = %q, want doubao-embedding-vision-251215", cfg.QwenEmbeddingModel)
	}
	if cfg.QwenEmbeddingDimension != 1024 {
		t.Fatalf("QwenEmbeddingDimension = %d, want 1024", cfg.QwenEmbeddingDimension)
	}
}

func TestLoadUsesVolcengineAPIKeyAsProviderFallback(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("QWEN_EMBEDDING_API_KEY", "")
	t.Setenv("VOLCENGINE_API_KEY", "volcengine-key")

	cfg := Load()

	if cfg.DeepSeekAPIKey != "volcengine-key" {
		t.Fatalf("DeepSeekAPIKey = %q, want Volcengine API key", cfg.DeepSeekAPIKey)
	}
	if cfg.QwenEmbeddingAPIKey != "volcengine-key" {
		t.Fatalf("QwenEmbeddingAPIKey = %q, want Volcengine API key", cfg.QwenEmbeddingAPIKey)
	}
}
