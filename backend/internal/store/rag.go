package store

import (
	"context"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoStore) SearchChunks(ctx context.Context, kbIDs []primitive.ObjectID, limit int64) ([]models.Chunk, error) {
	query := bson.M{}
	if len(kbIDs) > 0 {
		query["knowledgeBaseId"] = bson.M{"$in": kbIDs}
	}
	cursor, err := s.db.Collection("chunks").Find(ctx, query, options.Find().SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var chunks []models.Chunk
	if err := cursor.All(ctx, &chunks); err != nil {
		return nil, err
	}
	return chunks, nil
}

func (s *MongoStore) GetChunksByVectorIDs(ctx context.Context, vectorIDs []string) ([]models.Chunk, error) {
	if len(vectorIDs) == 0 {
		return nil, nil
	}
	cursor, err := s.db.Collection("chunks").Find(ctx, bson.M{"vectorId": bson.M{"$in": vectorIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var chunks []models.Chunk
	if err := cursor.All(ctx, &chunks); err != nil {
		return nil, err
	}
	return chunks, nil
}
