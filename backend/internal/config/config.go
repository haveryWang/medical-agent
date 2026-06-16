package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultQwenEmbeddingModel     = "doubao-embedding-vision-251215"
	DefaultQwenEmbeddingDimension = 2048
)

type Config struct {
	HTTPAddr               string
	JWTSecret              string
	SessionTTL             time.Duration
	MongoURI               string
	MongoDatabase          string
	UploadDir              string
	MaxUploadBytes         int64
	QdrantURL              string
	QdrantCollection       string
	DeepSeekBaseURL        string
	DeepSeekAPIKey         string
	DeepSeekChatModel      string
	QwenEmbeddingBaseURL   string
	QwenEmbeddingAPIKey    string
	QwenEmbeddingModel     string
	QwenEmbeddingDimension int
	RetrievalTopK          int
}

func Load() Config {
	volcengineAPIKey := os.Getenv("VOLCENGINE_API_KEY")
	cfg := Config{
		HTTPAddr:               getenv("HTTP_ADDR", ":8080"),
		JWTSecret:              getenv("JWT_SECRET", "local-dev-change-me"),
		SessionTTL:             time.Duration(getenvInt("SESSION_TTL_HOURS", 24)) * time.Hour,
		MongoURI:               getenv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:          getenv("MONGODB_DATABASE", "medical_agent"),
		UploadDir:              getenv("UPLOAD_DIR", "../data/uploads"),
		MaxUploadBytes:         int64(getenvInt("MAX_UPLOAD_MB", 15)) * 1024 * 1024,
		QdrantURL:              strings.TrimRight(getenv("QDRANT_URL", "http://localhost:6333"), "/"),
		QdrantCollection:       getenv("QDRANT_COLLECTION", "medical_agent_chunks"),
		DeepSeekBaseURL:        strings.TrimRight(getenv("DEEPSEEK_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"), "/"),
		DeepSeekAPIKey:         getenv("DEEPSEEK_API_KEY", volcengineAPIKey),
		DeepSeekChatModel:      getenv("DEEPSEEK_CHAT_MODEL", "DeepSeek-V4-flash"),
		QwenEmbeddingBaseURL:   strings.TrimRight(getenv("QWEN_EMBEDDING_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"), "/"),
		QwenEmbeddingAPIKey:    getenv("QWEN_EMBEDDING_API_KEY", volcengineAPIKey),
		QwenEmbeddingModel:     getenv("QWEN_EMBEDDING_MODEL", DefaultQwenEmbeddingModel),
		QwenEmbeddingDimension: getenvInt("QWEN_EMBEDDING_DIMENSION", DefaultQwenEmbeddingDimension),
		RetrievalTopK:          getenvInt("RETRIEVAL_TOP_K", 5),
	}
	if cfg.QwenEmbeddingModel == DefaultQwenEmbeddingModel {
		cfg.QwenEmbeddingDimension = DefaultQwenEmbeddingDimension
	}
	return cfg
}

func getenv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
