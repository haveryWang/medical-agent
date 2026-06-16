package store

import (
	"context"
	"strings"
	"time"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoStore) GetModelConfig(ctx context.Context, fallback config.Config) (models.ModelConfig, error) {
	var current models.ModelConfig
	err := s.db.Collection("model_configs").FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.D{{Key: "updatedAt", Value: -1}})).Decode(&current)
	if err != nil && err != mongo.ErrNoDocuments {
		return models.ModelConfig{}, err
	}
	return mergeModelConfig(current, fallback), nil
}

func (s *MongoStore) EnsureModelConfig(ctx context.Context, fallback config.Config) error {
	collection := s.db.Collection("model_configs")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	if count == 0 {
		_, err = collection.InsertOne(ctx, defaultModelConfig(fallback))
		return err
	}
	_, err = collection.UpdateMany(ctx, legacyDefaultModelConfigFilter(), bson.M{"$set": bson.M{
		"deepSeekBaseUrl":        fallback.DeepSeekBaseURL,
		"deepSeekChatModel":      fallback.DeepSeekChatModel,
		"qwenEmbeddingBaseUrl":   fallback.QwenEmbeddingBaseURL,
		"qwenEmbeddingModel":     fallback.QwenEmbeddingModel,
		"qwenEmbeddingDimension": fallback.QwenEmbeddingDimension,
		"updatedAt":              time.Now(),
	}})
	return err
}

func (s *MongoStore) SaveModelConfig(ctx context.Context, next models.ModelConfig, fallback config.Config) (models.ModelConfig, error) {
	current, err := s.GetModelConfig(ctx, fallback)
	if err != nil {
		return models.ModelConfig{}, err
	}
	if strings.TrimSpace(next.DeepSeekAPIKey) == "" {
		next.DeepSeekAPIKey = current.DeepSeekAPIKey
	}
	if strings.TrimSpace(next.QwenEmbeddingAPIKey) == "" {
		next.QwenEmbeddingAPIKey = current.QwenEmbeddingAPIKey
	}
	next = mergeModelConfig(next, fallback)
	next.ID = current.ID
	next.UpdatedAt = time.Now()

	collection := s.db.Collection("model_configs")
	if next.ID.IsZero() {
		_, err = collection.InsertOne(ctx, next)
		return next, err
	}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": next.ID}, bson.M{"$set": bson.M{
		"deepSeekBaseUrl":        next.DeepSeekBaseURL,
		"deepSeekAPIKey":         next.DeepSeekAPIKey,
		"deepSeekChatModel":      next.DeepSeekChatModel,
		"qwenEmbeddingBaseUrl":   next.QwenEmbeddingBaseURL,
		"qwenEmbeddingAPIKey":    next.QwenEmbeddingAPIKey,
		"qwenEmbeddingModel":     next.QwenEmbeddingModel,
		"qwenEmbeddingDimension": next.QwenEmbeddingDimension,
		"updatedAt":              next.UpdatedAt,
	}})
	return next, err
}

func defaultModelConfig(fallback config.Config) models.ModelConfig {
	modelConfig := mergeModelConfig(models.ModelConfig{}, fallback)
	modelConfig.UpdatedAt = time.Now()
	return modelConfig
}

func legacyDefaultModelConfigFilter() bson.M {
	return bson.M{"$or": []bson.M{
		{"deepSeekBaseUrl": bson.M{"$in": []string{"", "https://api.deepseek.com"}}},
		{"deepSeekChatModel": bson.M{"$in": []string{"", "deepseek-v4-flash-260425"}}},
		{"qwenEmbeddingBaseUrl": bson.M{"$in": []string{"", "https://dashscope.aliyuncs.com/compatible-mode/v1"}}},
		{"qwenEmbeddingModel": bson.M{"$in": []string{"", "Qwen3-Embedding"}}},
		{"qwenEmbeddingModel": "doubao-embedding-vision-251215", "qwenEmbeddingDimension": bson.M{"$ne": 2048}},
	}}
}

func mergeModelConfig(current models.ModelConfig, fallback config.Config) models.ModelConfig {
	if strings.TrimSpace(current.DeepSeekBaseURL) == "" {
		current.DeepSeekBaseURL = fallback.DeepSeekBaseURL
	} else {
		current.DeepSeekBaseURL = strings.TrimRight(strings.TrimSpace(current.DeepSeekBaseURL), "/")
	}
	if strings.TrimSpace(current.DeepSeekAPIKey) == "" {
		current.DeepSeekAPIKey = fallback.DeepSeekAPIKey
	}
	if strings.TrimSpace(current.DeepSeekChatModel) == "" {
		current.DeepSeekChatModel = fallback.DeepSeekChatModel
	}
	if strings.TrimSpace(current.QwenEmbeddingBaseURL) == "" {
		current.QwenEmbeddingBaseURL = fallback.QwenEmbeddingBaseURL
	} else {
		current.QwenEmbeddingBaseURL = strings.TrimRight(strings.TrimSpace(current.QwenEmbeddingBaseURL), "/")
	}
	if strings.TrimSpace(current.QwenEmbeddingAPIKey) == "" {
		current.QwenEmbeddingAPIKey = fallback.QwenEmbeddingAPIKey
	}
	if strings.TrimSpace(current.QwenEmbeddingModel) == "" {
		current.QwenEmbeddingModel = fallback.QwenEmbeddingModel
	}
	if current.QwenEmbeddingDimension <= 0 {
		current.QwenEmbeddingDimension = fallback.QwenEmbeddingDimension
	}
	if current.QwenEmbeddingModel == config.DefaultQwenEmbeddingModel {
		current.QwenEmbeddingDimension = config.DefaultQwenEmbeddingDimension
	}
	return current
}
