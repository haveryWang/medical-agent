package store

import (
	"context"
	"testing"
	"time"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestPendingJobsClaimsJobsAtomically(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("uses findAndModify to claim pending jobs", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		jobID := primitive.NewObjectID()
		kbID := primitive.NewObjectID()
		docID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(
				bson.E{Key: "value", Value: bson.D{
					{Key: "_id", Value: jobID},
					{Key: "knowledgeBaseId", Value: kbID},
					{Key: "documentId", Value: docID},
					{Key: "status", Value: "processing"},
				}},
			),
			mtest.CreateSuccessResponse(),
		)

		jobs, err := store.PendingJobs(context.Background(), 1)
		if err != nil {
			mt.Fatalf("PendingJobs error: %v", err)
		}
		if len(jobs) != 1 {
			mt.Fatalf("jobs = %#v, want one claimed job", jobs)
		}
		if jobs[0].ID != jobID {
			mt.Fatalf("job id = %v, want %v", jobs[0].ID, jobID)
		}

		event := mt.GetStartedEvent()
		if event == nil {
			mt.Fatal("expected a Mongo command")
		}
		if event.CommandName != "findAndModify" {
			mt.Fatalf("command = %q, want findAndModify", event.CommandName)
		}
		if got := event.Command.Lookup("query", "status").StringValue(); got != "pending" {
			mt.Fatalf("claim query status = %q, want pending", got)
		}
		if got := event.Command.Lookup("update", "$set", "status").StringValue(); got != "processing" {
			mt.Fatalf("claim update status = %q, want processing", got)
		}
	})

	mt.Run("stops when no pending job remains", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		firstID := primitive.NewObjectID()
		secondID := primitive.NewObjectID()
		kbID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(bson.E{Key: "value", Value: bson.D{
				{Key: "_id", Value: firstID},
				{Key: "knowledgeBaseId", Value: kbID},
				{Key: "documentId", Value: primitive.NewObjectID()},
				{Key: "status", Value: "processing"},
			}}),
			mtest.CreateSuccessResponse(bson.E{Key: "value", Value: bson.D{
				{Key: "_id", Value: secondID},
				{Key: "knowledgeBaseId", Value: kbID},
				{Key: "documentId", Value: primitive.NewObjectID()},
				{Key: "status", Value: "processing"},
			}}),
			mtest.CreateSuccessResponse(),
		)

		jobs, err := store.PendingJobs(context.Background(), 5)
		if err != nil {
			mt.Fatalf("PendingJobs error: %v", err)
		}
		if len(jobs) != 2 {
			mt.Fatalf("jobs = %#v, want two claimed jobs", jobs)
		}
		if jobs[0].ID != firstID || jobs[1].ID != secondID {
			mt.Fatalf("job ids = %v, %v; want %v, %v", jobs[0].ID, jobs[1].ID, firstID, secondID)
		}
	})
}

func TestCompleteJobWithChunksReplacesDocumentChunks(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("deletes stale chunks before inserting rebuilt chunks", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		docID := primitive.NewObjectID()
		job := models.IngestionJob{
			ID:              primitive.NewObjectID(),
			KnowledgeBaseID: primitive.NewObjectID(),
			DocumentID:      docID,
			Status:          "processing",
		}
		chunks := []models.Chunk{{
			ID:              primitive.NewObjectID(),
			KnowledgeBaseID: job.KnowledgeBaseID,
			DocumentID:      docID,
			Text:            "新分片",
			ChunkIndex:      0,
			CreatedAt:       time.Now(),
		}}
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(3)}),
			mtest.CreateSuccessResponse(bson.E{Key: "insertedIds", Value: bson.A{chunks[0].ID}}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}),
			mtest.CreateCursorResponse(0, "test.documents", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(1)}}),
			mtest.CreateCursorResponse(0, "test.chunks", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(1)}}),
			mtest.CreateCursorResponse(0, "test.ingestion_jobs", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(0)}}),
			mtest.CreateCursorResponse(0, "test.documents", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(0)}}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}),
		)

		if err := store.CompleteJobWithChunks(context.Background(), job, chunks); err != nil {
			mt.Fatalf("CompleteJobWithChunks error: %v", err)
		}

		first := mt.GetStartedEvent()
		second := mt.GetStartedEvent()
		third := mt.GetStartedEvent()
		if first.CommandName != "update" {
			mt.Fatalf("first command = %q, want update", first.CommandName)
		}
		if second.CommandName != "delete" {
			mt.Fatalf("second command = %q, want delete stale chunks", second.CommandName)
		}
		if third.CommandName != "insert" {
			mt.Fatalf("third command = %q, want insert rebuilt chunks", third.CommandName)
		}
		fourth := mt.GetStartedEvent()
		fifth := mt.GetStartedEvent()
		if fourth.CommandName != "update" {
			mt.Fatalf("fourth command = %q, want document update", fourth.CommandName)
		}
		if fifth.CommandName != "update" {
			mt.Fatalf("fifth command = %q, want job completion update", fifth.CommandName)
		}
		if got := fifth.Command.Lookup("updates").Array().Index(0).Value().Document().Lookup("u", "$set", "status").StringValue(); got != "completed" {
			mt.Fatalf("final job status = %q, want completed", got)
		}
	})
}
