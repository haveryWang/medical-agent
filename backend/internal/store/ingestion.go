package store

import (
	"context"
	"time"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeletedDocumentResult struct {
	DocumentID      primitive.ObjectID
	KnowledgeBaseID primitive.ObjectID
	VectorIDs       []string
	DeletedChunks   int64
}

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
	job, err := s.GetIngestionJob(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = s.db.Collection("ingestion_jobs").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "pending", "step": "retry", "error": "", "updatedAt": now}})
	if err != nil {
		return err
	}
	_, err = s.db.Collection("documents").UpdateOne(ctx, bson.M{"_id": job.DocumentID}, bson.M{"$set": bson.M{"status": "pending", "failureReason": "", "updatedAt": now}})
	return err
}

func (s *MongoStore) CompleteJobWithChunks(ctx context.Context, job models.IngestionJob, chunks []models.Chunk) error {
	now := time.Now()
	updateResult, err := s.db.Collection("ingestion_jobs").UpdateOne(
		ctx,
		bson.M{"_id": job.ID, "status": "processing"},
		bson.M{"$set": bson.M{"status": "indexing", "step": "indexing", "updatedAt": now}},
	)
	if err != nil {
		return err
	}
	if updateResult.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	if _, err := s.db.Collection("chunks").DeleteMany(ctx, bson.M{"documentId": job.DocumentID}); err != nil {
		return err
	}
	if len(chunks) > 0 {
		records := make([]any, 0, len(chunks))
		for _, chunk := range chunks {
			records = append(records, chunk)
		}
		if _, err := s.db.Collection("chunks").InsertMany(ctx, records); err != nil {
			return err
		}
	}
	_, err = s.db.Collection("documents").UpdateOne(ctx, bson.M{"_id": job.DocumentID}, bson.M{"$set": bson.M{"status": "completed", "updatedAt": now}})
	if err != nil {
		return err
	}
	if _, err = s.db.Collection("ingestion_jobs").UpdateOne(
		ctx,
		bson.M{"_id": job.ID, "status": "indexing"},
		bson.M{"$set": bson.M{"status": "completed", "step": "indexed", "updatedAt": now}},
	); err != nil {
		return err
	}
	return s.recountKnowledgeBase(ctx, job.KnowledgeBaseID)
}

func (s *MongoStore) FailJob(ctx context.Context, job models.IngestionJob, message string) error {
	now := time.Now()
	updateResult, err := s.db.Collection("ingestion_jobs").UpdateOne(ctx, bson.M{"_id": job.ID, "status": bson.M{"$in": []string{"processing", "indexing"}}}, bson.M{"$set": bson.M{"status": "failed", "step": "failed", "error": message, "updatedAt": now}, "$inc": bson.M{"attempts": 1}})
	if err != nil {
		return err
	}
	if updateResult.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	if _, err = s.db.Collection("documents").UpdateOne(ctx, bson.M{"_id": job.DocumentID}, bson.M{"$set": bson.M{"status": "failed", "failureReason": message, "updatedAt": now}}); err != nil {
		return err
	}
	return s.recountKnowledgeBase(ctx, job.KnowledgeBaseID)
}

func (s *MongoStore) PendingJobs(ctx context.Context, limit int64) ([]models.IngestionJob, error) {
	if limit <= 0 {
		limit = 1
	}
	jobs := make([]models.IngestionJob, 0, limit)
	for int64(len(jobs)) < limit {
		now := time.Now()
		var job models.IngestionJob
		err := s.db.Collection("ingestion_jobs").FindOneAndUpdate(
			ctx,
			bson.M{"status": "pending"},
			bson.M{"$set": bson.M{"status": "processing", "step": "processing", "updatedAt": now}},
			options.FindOneAndUpdate().SetSort(bson.D{{Key: "createdAt", Value: 1}}).SetReturnDocument(options.After),
		).Decode(&job)
		if err == mongo.ErrNoDocuments {
			break
		}
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s *MongoStore) GetDocument(ctx context.Context, id primitive.ObjectID) (models.Document, error) {
	var doc models.Document
	err := s.db.Collection("documents").FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	return doc, err
}

func (s *MongoStore) ListDocumentChunks(ctx context.Context, documentID primitive.ObjectID) ([]models.Chunk, error) {
	cursor, err := s.db.Collection("chunks").Find(
		ctx,
		bson.M{"documentId": documentID},
		options.Find().SetSort(bson.D{{Key: "chunkIndex", Value: 1}}),
	)
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

func (s *MongoStore) DeleteDocumentCascade(ctx context.Context, doc models.Document) (DeletedDocumentResult, error) {
	chunks, err := s.ListDocumentChunks(ctx, doc.ID)
	if err != nil {
		return DeletedDocumentResult{}, err
	}
	vectorIDs := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk.VectorID != "" {
			vectorIDs = append(vectorIDs, chunk.VectorID)
		}
	}

	result := DeletedDocumentResult{
		DocumentID:      doc.ID,
		KnowledgeBaseID: doc.KnowledgeBaseID,
		VectorIDs:       vectorIDs,
		DeletedChunks:   int64(len(chunks)),
	}
	deleteResult, err := s.db.Collection("documents").DeleteOne(ctx, bson.M{"_id": doc.ID, "knowledgeBaseId": doc.KnowledgeBaseID})
	if err != nil {
		return DeletedDocumentResult{}, err
	}
	if deleteResult.DeletedCount == 0 {
		return DeletedDocumentResult{}, mongo.ErrNoDocuments
	}
	if _, err = s.db.Collection("chunks").DeleteMany(ctx, bson.M{"documentId": doc.ID}); err != nil {
		return DeletedDocumentResult{}, err
	}
	if _, err = s.db.Collection("ingestion_jobs").DeleteMany(ctx, bson.M{"documentId": doc.ID}); err != nil {
		return DeletedDocumentResult{}, err
	}
	if err = s.recountKnowledgeBase(ctx, doc.KnowledgeBaseID); err != nil {
		return DeletedDocumentResult{}, err
	}
	return result, nil
}

func (s *MongoStore) recountKnowledgeBase(ctx context.Context, knowledgeBaseID primitive.ObjectID) error {
	docCount, err := s.db.Collection("documents").CountDocuments(ctx, bson.M{"knowledgeBaseId": knowledgeBaseID})
	if err != nil {
		return err
	}
	chunkCount, err := s.db.Collection("chunks").CountDocuments(ctx, bson.M{"knowledgeBaseId": knowledgeBaseID})
	if err != nil {
		return err
	}
	buildStatus, err := s.knowledgeBuildStatus(ctx, knowledgeBaseID)
	if err != nil {
		return err
	}
	_, err = s.db.Collection("knowledge_bases").UpdateOne(
		ctx,
		bson.M{"_id": knowledgeBaseID},
		bson.M{"$set": bson.M{
			"documentCount": docCount,
			"chunkCount":    chunkCount,
			"buildStatus":   buildStatus,
			"updatedAt":     time.Now(),
		}},
	)
	return err
}

func (s *MongoStore) knowledgeBuildStatus(ctx context.Context, knowledgeBaseID primitive.ObjectID) (string, error) {
	pendingJobs, err := s.db.Collection("ingestion_jobs").CountDocuments(ctx, bson.M{"knowledgeBaseId": knowledgeBaseID, "status": bson.M{"$in": []string{"pending", "processing", "indexing"}}})
	if err != nil {
		return "", err
	}
	if pendingJobs > 0 {
		return "building", nil
	}
	failedDocs, err := s.db.Collection("documents").CountDocuments(ctx, bson.M{"knowledgeBaseId": knowledgeBaseID, "status": "failed"})
	if err != nil {
		return "", err
	}
	if failedDocs > 0 {
		return "failed", nil
	}
	return "completed", nil
}
