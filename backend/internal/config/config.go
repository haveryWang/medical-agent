package config

import (
	"os"
	"strconv"
	"strings"
	"time"
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
	return Config{
		HTTPAddr:               getenv("HTTP_ADDR", ":8080"),
		JWTSecret:              getenv("JWT_SECRET", "local-dev-change-me"),
		SessionTTL:             time.Duration(getenvInt("SESSION_TTL_HOURS", 24)) * time.Hour,
		MongoURI:               getenv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase:          getenv("MONGODB_DATABASE", "medical_agent"),
		UploadDir:              getenv("UPLOAD_DIR", "../data/uploads"),
		MaxUploadBytes:         int64(getenvInt("MAX_UPLOAD_MB", 50)) * 1024 * 1024,
		QdrantURL:              strings.TrimRight(getenv("QDRANT_URL", "http://localhost:6333"), "/"),
		QdrantCollection:       getenv("QDRANT_COLLECTION", "medical_agent_chunks"),
		DeepSeekBaseURL:        strings.TrimRight(getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com"), "/"),
		DeepSeekAPIKey:         os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekChatModel:      getenv("DEEPSEEK_CHAT_MODEL", "deepseek-v4-flash"),
		QwenEmbeddingBaseURL:   strings.TrimRight(getenv("QWEN_EMBEDDING_BASE_URL", ""), "/"),
		QwenEmbeddingAPIKey:    os.Getenv("QWEN_EMBEDDING_API_KEY"),
		QwenEmbeddingModel:     getenv("QWEN_EMBEDDING_MODEL", "Qwen3-Embedding"),
		QwenEmbeddingDimension: getenvInt("QWEN_EMBEDDING_DIMENSION", 1024),
		RetrievalTopK:          getenvInt("RETRIEVAL_TOP_K", 5),
	}
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
