package store

import (
	"context"
	"testing"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestImportPolicyDocumentsUsesDedicatedCollections(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("inserts batch and policy documents outside knowledge collections", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		actorID := primitive.NewObjectID()
		docs := []models.PolicyDocument{{
			Title:    "国家医学中心建设通知",
			Summary:  "围绕医学中心建设提出重点任务",
			Date:     "2026-06-08",
			Category: "国家医学中心",
		}}
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(bson.E{Key: "insertedId", Value: primitive.NewObjectID()}),
			mtest.CreateSuccessResponse(bson.E{Key: "insertedIds", Value: bson.A{primitive.NewObjectID()}}),
		)

		batch, err := store.ImportPolicyDocuments(context.Background(), actorID, "政策文件.xlsx", docs, 1, 0, nil)
		if err != nil {
			mt.Fatalf("ImportPolicyDocuments error: %v", err)
		}
		if batch.ActorID != actorID || batch.Imported != 1 || batch.Skipped != 0 {
			mt.Fatalf("batch = %#v", batch)
		}

		first := mt.GetStartedEvent()
		second := mt.GetStartedEvent()
		if first.CommandName != "insert" || first.Command.Lookup("insert").StringValue() != "policy_import_batches" {
			mt.Fatalf("first command = %s %s, want insert policy_import_batches", first.CommandName, first.Command.Lookup("insert").StringValue())
		}
		if second.CommandName != "insert" || second.Command.Lookup("insert").StringValue() != "policy_documents" {
			mt.Fatalf("second command = %s %s, want insert policy_documents", second.CommandName, second.Command.Lookup("insert").StringValue())
		}
	})
}

func TestListPolicyDocumentsFiltersByCategory(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("queries category and returns policy documents", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		docID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.policy_documents", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(1)}}),
			mtest.CreateCursorResponse(0, "test.policy_documents", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: docID},
				{Key: "title", Value: "医保支付方式改革"},
				{Key: "summary", Value: "完善医保支付机制"},
				{Key: "date", Value: "2026-06-01"},
				{Key: "category", Value: "医保医药"},
			}),
		)

		result, err := store.ListPolicyDocuments(context.Background(), PolicyFilter{Category: "医保医药"})
		if err != nil {
			mt.Fatalf("ListPolicyDocuments error: %v", err)
		}
		if len(result.Items) != 1 || result.Items[0].ID != docID {
			mt.Fatalf("items = %#v, want policy %v", result.Items, docID)
		}

		_ = mt.GetStartedEvent()
		event := mt.GetStartedEvent()
		if event.CommandName != "find" || event.Command.Lookup("find").StringValue() != "policy_documents" {
			mt.Fatalf("command = %s %s, want find policy_documents", event.CommandName, event.Command.Lookup("find").StringValue())
		}
		if got := event.Command.Lookup("filter", "category").StringValue(); got != "医保医药" {
			mt.Fatalf("category filter = %q", got)
		}
	})
}

func TestListPolicyDocumentsFiltersByCategoryAndDate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("queries category and date together", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.policy_documents", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(0)}}),
			mtest.CreateCursorResponse(0, "test.policy_documents", mtest.FirstBatch),
		)

		_, err := store.ListPolicyDocuments(context.Background(), PolicyFilter{Category: "医保医药", Date: "2026-06"})
		if err != nil {
			mt.Fatalf("ListPolicyDocuments error: %v", err)
		}

		_ = mt.GetStartedEvent()
		find := mt.GetStartedEvent()
		if got := find.Command.Lookup("filter", "category").StringValue(); got != "医保医药" {
			mt.Fatalf("category filter = %q", got)
		}
		dateFilter := find.Command.Lookup("filter", "date")
		if got := dateFilter.Document().Lookup("$regex").StringValue(); got != "^2026-06" {
			mt.Fatalf("date regex = %q, want month prefix", got)
		}
	})
}

func TestListPolicyDocumentsUsesPageAndSize(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("counts total and applies skip limit", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		docID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.policy_documents", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(12)}}),
			mtest.CreateCursorResponse(0, "test.policy_documents", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: docID},
				{Key: "title", Value: "分页政策"},
				{Key: "summary", Value: "第二页"},
				{Key: "date", Value: "2026-06-01"},
				{Key: "category", Value: "医保医药"},
			}),
		)

		result, err := store.ListPolicyDocuments(context.Background(), PolicyFilter{Page: 2, PageSize: 5})
		if err != nil {
			mt.Fatalf("ListPolicyDocuments error: %v", err)
		}
		if result.Total != 12 || result.Page != 2 || result.PageSize != 5 || len(result.Items) != 1 {
			mt.Fatalf("result = %#v", result)
		}

		count := mt.GetStartedEvent()
		find := mt.GetStartedEvent()
		if count.CommandName != "aggregate" || count.Command.Lookup("aggregate").StringValue() != "policy_documents" {
			mt.Fatalf("first command = %s %s, want count policy_documents", count.CommandName, count.Command.Lookup("aggregate").StringValue())
		}
		if got := find.Command.Lookup("skip").Int64(); got != 5 {
			mt.Fatalf("skip = %d, want 5", got)
		}
		if got := find.Command.Lookup("limit").Int64(); got != 5 {
			mt.Fatalf("limit = %d, want 5", got)
		}
	})
}

func TestDeletePolicyDocumentDeletesSingleRecord(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("deletes from policy_documents by id", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		docID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}))

		if err := store.DeletePolicyDocument(context.Background(), docID); err != nil {
			mt.Fatalf("DeletePolicyDocument error: %v", err)
		}

		event := mt.GetStartedEvent()
		if event.CommandName != "delete" || event.Command.Lookup("delete").StringValue() != "policy_documents" {
			mt.Fatalf("command = %s %s, want delete policy_documents", event.CommandName, event.Command.Lookup("delete").StringValue())
		}
		if got := event.Command.Lookup("deletes").Array().Index(0).Value().Document().Lookup("q", "_id").ObjectID(); got != docID {
			mt.Fatalf("deleted id = %v, want %v", got, docID)
		}
	})
}

func TestListPolicyFacetsAggregatesCategoryAndMonth(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("aggregates facet counts from policy documents", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.policy_documents", mtest.FirstBatch, bson.D{
			{Key: "category", Value: bson.A{
				bson.D{{Key: "_id", Value: "医保医药"}, {Key: "count", Value: int32(3)}},
			}},
			{Key: "date", Value: bson.A{
				bson.D{{Key: "_id", Value: "2026-06"}, {Key: "count", Value: int32(2)}},
			}},
		}))

		facets, err := store.ListPolicyFacets(context.Background())
		if err != nil {
			mt.Fatalf("ListPolicyFacets error: %v", err)
		}
		if len(facets.Categories) != 1 || facets.Categories[0].Value != "医保医药" || facets.Categories[0].Count != 3 {
			mt.Fatalf("category facets = %#v", facets.Categories)
		}
		if len(facets.Dates) != 1 || facets.Dates[0].Value != "2026-06" || facets.Dates[0].Count != 2 {
			mt.Fatalf("date facets = %#v", facets.Dates)
		}

		event := mt.GetStartedEvent()
		if event.CommandName != "aggregate" || event.Command.Lookup("aggregate").StringValue() != "policy_documents" {
			mt.Fatalf("command = %s %s, want aggregate policy_documents", event.CommandName, event.Command.Lookup("aggregate").StringValue())
		}
	})
}

func TestPolicyDocumentModelHasDisplayFields(t *testing.T) {
	doc := models.PolicyDocument{Title: "标题", Summary: "摘要", Interpretation: "解读", Date: "2026-06-08", Category: "科技创新"}
	if doc.Title == "" || doc.Summary == "" || doc.Interpretation == "" || doc.Date == "" || doc.Category == "" {
		t.Fatalf("policy display fields missing: %#v", doc)
	}
}
