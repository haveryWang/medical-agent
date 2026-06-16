package store

import (
	"context"
	"time"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoStore) ListKnowledgeBases(ctx context.Context, filter KnowledgeFilter) (PagedKnowledgeBases, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 10
	}
	query := bson.M{"status": bson.M{"$ne": "deleted"}}
	if filter.Scenario != "" {
		query["scenario"] = filter.Scenario
	}
	if filter.Tag != "" {
		query["tags"] = filter.Tag
	}
	if filter.Department != "" {
		query["department"] = filter.Department
	}
	if filter.Keyword != "" {
		query["$or"] = []bson.M{
			{"name": bson.M{"$regex": filter.Keyword, "$options": "i"}},
			{"description": bson.M{"$regex": filter.Keyword, "$options": "i"}},
		}
	}
	total, err := s.db.Collection("knowledge_bases").CountDocuments(ctx, query)
	if err != nil {
		return PagedKnowledgeBases{}, err
	}
	opts := options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetSkip((filter.Page - 1) * filter.PageSize).SetLimit(filter.PageSize)
	cursor, err := s.db.Collection("knowledge_bases").Find(ctx, query, opts)
	if err != nil {
		return PagedKnowledgeBases{}, err
	}
	defer cursor.Close(ctx)
	var items []models.KnowledgeBase
	if err := cursor.All(ctx, &items); err != nil {
		return PagedKnowledgeBases{}, err
	}
	return PagedKnowledgeBases{Items: items, Total: total, Page: filter.Page, Size: filter.PageSize}, nil
}

func (s *MongoStore) GetKnowledgeBase(ctx context.Context, id primitive.ObjectID) (models.KnowledgeBase, error) {
	var kb models.KnowledgeBase
	err := s.db.Collection("knowledge_bases").FindOne(ctx, bson.M{"_id": id, "status": bson.M{"$ne": "deleted"}}).Decode(&kb)
	return kb, err
}

func (s *MongoStore) CreateKnowledgeBase(ctx context.Context, kb models.KnowledgeBase) (models.KnowledgeBase, error) {
	now := time.Now()
	kb.ID = primitive.NewObjectID()
	kb.Status = "active"
	kb.BuildStatus = "completed"
	kb.CreatedAt = now
	kb.UpdatedAt = now
	if kb.RetrievalTopK == 0 {
		kb.RetrievalTopK = s.cfg.RetrievalTopK
	}
	if kb.SimilarityFloor == 0 {
		kb.SimilarityFloor = 0.2
	}
	_, err := s.db.Collection("knowledge_bases").InsertOne(ctx, kb)
	return kb, err
}

func (s *MongoStore) UpdateKnowledgeBase(ctx context.Context, id primitive.ObjectID, patch bson.M) error {
	patch["updatedAt"] = time.Now()
	_, err := s.db.Collection("knowledge_bases").UpdateOne(ctx, bson.M{"_id": id, "status": bson.M{"$ne": "deleted"}}, bson.M{"$set": patch})
	return err
}

func (s *MongoStore) ListDocuments(ctx context.Context, knowledgeBaseID primitive.ObjectID) ([]models.Document, error) {
	cursor, err := s.db.Collection("documents").Find(ctx, bson.M{"knowledgeBaseId": knowledgeBaseID}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []models.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}
