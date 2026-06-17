package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	client *mongo.Client
	db     *mongo.Database
	cfg    config.Config
}

type KnowledgeFilter struct {
	Scenario   string
	Tag        string
	Department string
	Keyword    string
	Page       int64
	PageSize   int64
}

type PagedKnowledgeBases struct {
	Items []models.KnowledgeBase `json:"items"`
	Total int64                  `json:"total"`
	Page  int64                  `json:"page"`
	Size  int64                  `json:"size"`
}

func NewMongoStore(ctx context.Context, cfg config.Config) (*MongoStore, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	store := &MongoStore{client: client, db: client.Database(cfg.MongoDatabase), cfg: cfg}
	if err := store.EnsureIndexes(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	if err := store.EnsureModelConfig(ctx, cfg); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	if err := store.Seed(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return store, nil
}

func (s *MongoStore) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

func (s *MongoStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx, nil)
}

func (s *MongoStore) EnsureIndexes(ctx context.Context) error {
	indexes := map[string][]mongo.IndexModel{
		"users": {
			{Keys: bson.D{{Key: "account", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "status", Value: 1}}},
		},
		"sessions": {
			{Keys: bson.D{{Key: "tokenHash", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "userId", Value: 1}}},
			{Keys: bson.D{{Key: "expiresAt", Value: 1}}},
		},
		"knowledge_bases": {
			{Keys: bson.D{{Key: "name", Value: "text"}, {Key: "description", Value: "text"}}},
			{Keys: bson.D{{Key: "scenario", Value: 1}}},
			{Keys: bson.D{{Key: "tags", Value: 1}}},
			{Keys: bson.D{{Key: "department", Value: 1}}},
			{Keys: bson.D{{Key: "buildStatus", Value: 1}}},
			{Keys: bson.D{{Key: "updatedAt", Value: -1}}},
		},
		"documents": {
			{Keys: bson.D{{Key: "knowledgeBaseId", Value: 1}}},
			{Keys: bson.D{{Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "createdAt", Value: -1}}},
		},
		"chunks": {
			{Keys: bson.D{{Key: "knowledgeBaseId", Value: 1}}},
			{Keys: bson.D{{Key: "documentId", Value: 1}}},
			{Keys: bson.D{{Key: "vectorId", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"ingestion_jobs": {
			{Keys: bson.D{{Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "documentId", Value: 1}}},
			{Keys: bson.D{{Key: "updatedAt", Value: -1}}},
		},
		"conversations": {
			{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "updatedAt", Value: -1}}},
			{Keys: bson.D{{Key: "title", Value: "text"}}},
		},
		"messages": {
			{Keys: bson.D{{Key: "conversationId", Value: 1}, {Key: "createdAt", Value: 1}}},
			{Keys: bson.D{{Key: "role", Value: 1}}},
		},
		"audit_logs": {
			{Keys: bson.D{{Key: "actorId", Value: 1}, {Key: "createdAt", Value: -1}}},
			{Keys: bson.D{{Key: "action", Value: 1}}},
		},
		"review_notes": {
			{Keys: bson.D{{Key: "actorId", Value: 1}, {Key: "createdAt", Value: -1}}},
			{Keys: bson.D{{Key: "exported", Value: 1}, {Key: "createdAt", Value: 1}}},
			{Keys: bson.D{{Key: "exportBatchId", Value: 1}}},
		},
		"review_note_exports": {
			{Keys: bson.D{{Key: "actorId", Value: 1}, {Key: "createdAt", Value: -1}}},
		},
		"policy_documents": {
			{Keys: bson.D{{Key: "category", Value: 1}, {Key: "date", Value: -1}}},
			{Keys: bson.D{{Key: "title", Value: "text"}, {Key: "summary", Value: "text"}}},
			{Keys: bson.D{{Key: "importBatchId", Value: 1}}},
			{Keys: bson.D{{Key: "rowChecksum", Value: 1}}},
		},
		"policy_import_batches": {
			{Keys: bson.D{{Key: "actorId", Value: 1}, {Key: "createdAt", Value: -1}}},
		},
		"model_configs": {
			{Keys: bson.D{{Key: "updatedAt", Value: -1}}},
			{Keys: bson.D{{Key: "deepSeekChatModel", Value: 1}}},
			{Keys: bson.D{{Key: "qwenEmbeddingModel", Value: 1}}},
		},
	}
	for collection, models := range indexes {
		if _, err := s.db.Collection(collection).Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("create indexes for %s: %w", collection, err)
		}
	}
	return nil
}

func ObjectIDFromHex(value string) (primitive.ObjectID, error) {
	if strings.TrimSpace(value) == "" {
		return primitive.NilObjectID, errors.New("id is required")
	}
	return primitive.ObjectIDFromHex(value)
}
