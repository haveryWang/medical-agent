package store

import (
	"context"
	"time"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *MongoStore) CreateDocumentAndJob(ctx context.Context, doc models.Document) (models.Document, models.IngestionJob, error) {
	now := time.Now()
	doc.ID = primitive.NewObjectID()
	doc.Status = "pending"
	doc.CreatedAt = now
	doc.UpdatedAt = now
	job := models.IngestionJob{
		ID:              primitive.NewObjectID(),
		KnowledgeBaseID: doc.KnowledgeBaseID,
		DocumentID:      doc.ID,
		Status:          "pending",
		Step:            "uploaded",
		Attempts:        0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err := s.db.Collection("documents").InsertOne(ctx, doc)
	if err != nil {
		return doc, job, err
	}
	_, err = s.db.Collection("ingestion_jobs").InsertOne(ctx, job)
	if err != nil {
		return doc, job, err
	}
	_, _ = s.db.Collection("knowledge_bases").UpdateOne(ctx, bson.M{"_id": doc.KnowledgeBaseID}, bson.M{"$inc": bson.M{"documentCount": 1}, "$set": bson.M{"buildStatus": "building", "updatedAt": now}})
	return doc, job, nil
}

func (s *MongoStore) GetIngestionJob(ctx context.Context, id primitive.ObjectID) (models.IngestionJob, error) {
	var job models.IngestionJob
	err := s.db.Collection("ingestion_jobs").FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	return job, err
}

func (s *MongoStore) RetryIngestionJob(ctx context.Context, id primitive.ObjectID) error {
	_, err := s.db.Collection("ingestion_jobs").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "pending", "step": "retry", "error": "", "updatedAt": time.Now()}})
	return err
}

func (s *MongoStore) CompleteJobWithChunks(ctx context.Context, job models.IngestionJob, chunks []models.Chunk) error {
	now := time.Now()
	if len(chunks) > 0 {
		records := make([]any, 0, len(chunks))
		for _, chunk := range chunks {
			records = append(records, chunk)
		}
		if _, err := s.db.Collection("chunks").InsertMany(ctx, records); err != nil {
			return err
		}
	}
	_, err := s.db.Collection("ingestion_jobs").UpdateOne(ctx, bson.M{"_id": job.ID}, bson.M{"$set": bson.M{"status": "completed", "step": "indexed", "updatedAt": now}})
	if err != nil {
		return err
	}
	_, err = s.db.Collection("documents").UpdateOne(ctx, bson.M{"_id": job.DocumentID}, bson.M{"$set": bson.M{"status": "completed", "updatedAt": now}})
	if err != nil {
		return err
	}
	_, err = s.db.Collection("knowledge_bases").UpdateOne(ctx, bson.M{"_id": job.KnowledgeBaseID}, bson.M{"$inc": bson.M{"chunkCount": len(chunks)}, "$set": bson.M{"buildStatus": "completed", "updatedAt": now}})
	return err
}

func (s *MongoStore) FailJob(ctx context.Context, job models.IngestionJob, message string) error {
	now := time.Now()
	_, err := s.db.Collection("ingestion_jobs").UpdateOne(ctx, bson.M{"_id": job.ID}, bson.M{"$set": bson.M{"status": "failed", "step": "failed", "error": message, "updatedAt": now}, "$inc": bson.M{"attempts": 1}})
	if err != nil {
		return err
	}
	_, err = s.db.Collection("documents").UpdateOne(ctx, bson.M{"_id": job.DocumentID}, bson.M{"$set": bson.M{"status": "failed", "failureReason": message, "updatedAt": now}})
	return err
}

func (s *MongoStore) PendingJobs(ctx context.Context, limit int64) ([]models.IngestionJob, error) {
	cursor, err := s.db.Collection("ingestion_jobs").Find(ctx, bson.M{"status": "pending"}, options.Find().SetLimit(limit).SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var jobs []models.IngestionJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (s *MongoStore) GetDocument(ctx context.Context, id primitive.ObjectID) (models.Document, error) {
	var doc models.Document
	err := s.db.Collection("documents").FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	return doc, err
}
