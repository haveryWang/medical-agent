package store

import (
	"context"
	"errors"
	"regexp"
	"time"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PolicyFilter struct {
	Category string
	Date     string
	Keyword  string
	Limit    int64
	Page     int64
	PageSize int64
}

type PolicyFacetItem struct {
	Value string `bson:"value" json:"value"`
	Count int    `bson:"count" json:"count"`
}

type PolicyFacets struct {
	Categories []PolicyFacetItem `json:"categories"`
	Dates      []PolicyFacetItem `json:"dates"`
}

type PolicyListResult struct {
	Items    []models.PolicyDocument `json:"items"`
	Total    int64                   `json:"total"`
	Page     int64                   `json:"page"`
	PageSize int64                   `json:"size"`
}

func (s *MongoStore) ImportPolicyDocuments(ctx context.Context, actorID primitive.ObjectID, filename string, docs []models.PolicyDocument, imported int, skipped int, errors []string) (models.PolicyImportBatch, error) {
	now := time.Now()
	batch := models.PolicyImportBatch{
		ID:        primitive.NewObjectID(),
		ActorID:   actorID,
		Filename:  filename,
		Imported:  imported,
		Skipped:   skipped,
		Errors:    errors,
		CreatedAt: now,
	}
	if _, err := s.db.Collection("policy_import_batches").InsertOne(ctx, batch); err != nil {
		return models.PolicyImportBatch{}, err
	}
	if len(docs) == 0 {
		return batch, nil
	}
	values := make([]any, 0, len(docs))
	for _, doc := range docs {
		doc.ID = primitive.NewObjectID()
		doc.ImportBatchID = batch.ID
		doc.CreatedAt = now
		doc.UpdatedAt = now
		values = append(values, doc)
	}
	if _, err := s.db.Collection("policy_documents").InsertMany(ctx, values); err != nil {
		return models.PolicyImportBatch{}, err
	}
	return batch, nil
}

func (s *MongoStore) ListPolicyDocuments(ctx context.Context, filter PolicyFilter) (PolicyListResult, error) {
	query := bson.M{}
	if filter.Category != "" {
		query["category"] = filter.Category
	}
	if filter.Date != "" {
		query["date"] = bson.M{"$regex": "^" + regexp.QuoteMeta(filter.Date)}
	}
	if filter.Keyword != "" {
		query["title"] = bson.M{"$regex": regexp.QuoteMeta(filter.Keyword), "$options": "i"}
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = filter.Limit
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	total, err := s.db.Collection("policy_documents").CountDocuments(ctx, query)
	if err != nil {
		return PolicyListResult{}, err
	}
	cursor, err := s.db.Collection("policy_documents").Find(ctx, query, options.Find().
		SetSort(bson.D{{Key: "date", Value: -1}, {Key: "createdAt", Value: -1}}).
		SetSkip((page-1)*pageSize).
		SetLimit(pageSize))
	if err != nil {
		return PolicyListResult{}, err
	}
	defer cursor.Close(ctx)
	var items []models.PolicyDocument
	if err := cursor.All(ctx, &items); err != nil {
		return PolicyListResult{}, err
	}
	return PolicyListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *MongoStore) DeletePolicyDocument(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.db.Collection("policy_documents").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("政策记录不存在")
	}
	return nil
}

func (s *MongoStore) ListPolicyFacets(ctx context.Context) (PolicyFacets, error) {
	pipeline := []bson.M{
		{"$facet": bson.M{
			"category": []bson.M{
				{"$match": bson.M{"category": bson.M{"$ne": ""}}},
				{"$group": bson.M{"_id": "$category", "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"count": -1, "_id": 1}},
			},
			"date": []bson.M{
				{"$match": bson.M{"date": bson.M{"$ne": ""}}},
				{"$project": bson.M{"month": bson.M{"$substrCP": []any{"$date", 0, 7}}}},
				{"$match": bson.M{"month": bson.M{"$regex": `^\d{4}-\d{2}$`}}},
				{"$group": bson.M{"_id": "$month", "count": bson.M{"$sum": 1}}},
				{"$sort": bson.M{"_id": -1}},
			},
		}},
	}
	cursor, err := s.db.Collection("policy_documents").Aggregate(ctx, pipeline)
	if err != nil {
		return PolicyFacets{}, err
	}
	defer cursor.Close(ctx)
	var raw []struct {
		Category []struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		} `bson:"category"`
		Date []struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		} `bson:"date"`
	}
	if err := cursor.All(ctx, &raw); err != nil {
		return PolicyFacets{}, err
	}
	if len(raw) == 0 {
		return PolicyFacets{}, nil
	}
	result := PolicyFacets{
		Categories: make([]PolicyFacetItem, 0, len(raw[0].Category)),
		Dates:      make([]PolicyFacetItem, 0, len(raw[0].Date)),
	}
	for _, item := range raw[0].Category {
		result.Categories = append(result.Categories, PolicyFacetItem{Value: item.ID, Count: item.Count})
	}
	for _, item := range raw[0].Date {
		result.Dates = append(result.Dates, PolicyFacetItem{Value: item.ID, Count: item.Count})
	}
	return result, nil
}
