package store

import (
	"context"
	"strings"
	"testing"

	"medical-agent/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCreateReviewNoteStoresDedicatedRecord(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("inserts into review_notes only", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		actorID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "insertedId", Value: primitive.NewObjectID()}))

		note, err := store.CreateReviewNote(context.Background(), actorID, "  复盘内容  ")
		if err != nil {
			mt.Fatalf("CreateReviewNote error: %v", err)
		}
		if note.ActorID != actorID {
			mt.Fatalf("actor = %v, want %v", note.ActorID, actorID)
		}
		if note.Content != "复盘内容" {
			mt.Fatalf("content = %q", note.Content)
		}
		if note.Exported {
			mt.Fatal("new note should not be exported")
		}
		if note.CreatedAt.IsZero() {
			mt.Fatal("expected created time")
		}

		event := mt.GetStartedEvent()
		if event.CommandName != "insert" {
			mt.Fatalf("command = %q, want insert", event.CommandName)
		}
		if got := event.Command.Lookup("insert").StringValue(); got != "review_notes" {
			mt.Fatalf("collection = %q, want review_notes", got)
		}
	})
}

func TestClaimUnexportedReviewNotesCreatesBatchAndMarksNotes(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("reads unexported notes then writes export batch", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		actorID := primitive.NewObjectID()
		noteID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.review_notes", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: noteID},
				{Key: "actorId", Value: actorID},
				{Key: "content", Value: "复盘内容"},
				{Key: "exported", Value: false},
			}),
			mtest.CreateSuccessResponse(bson.E{Key: "insertedId", Value: primitive.NewObjectID()}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}),
		)

		notes, batch, err := store.ClaimUnexportedReviewNotes(context.Background(), actorID, "复盘笔记.md")
		if err != nil {
			mt.Fatalf("ClaimUnexportedReviewNotes error: %v", err)
		}
		if len(notes) != 1 || notes[0].ID != noteID {
			mt.Fatalf("notes = %#v, want note %v", notes, noteID)
		}
		if batch.ActorID != actorID || batch.NoteCount != 1 || batch.Filename != "复盘笔记.md" {
			mt.Fatalf("batch = %#v", batch)
		}

		find := mt.GetStartedEvent()
		insert := mt.GetStartedEvent()
		update := mt.GetStartedEvent()
		if find.CommandName != "find" || find.Command.Lookup("find").StringValue() != "review_notes" {
			mt.Fatalf("first command = %s %s, want find review_notes", find.CommandName, find.Command.Lookup("find").StringValue())
		}
		if insert.CommandName != "insert" || insert.Command.Lookup("insert").StringValue() != "review_note_exports" {
			mt.Fatalf("second command = %s %s, want insert review_note_exports", insert.CommandName, insert.Command.Lookup("insert").StringValue())
		}
		if update.CommandName != "update" || update.Command.Lookup("update").StringValue() != "review_notes" {
			mt.Fatalf("third command = %s %s, want update review_notes", update.CommandName, update.Command.Lookup("update").StringValue())
		}
		if got := update.Command.Lookup("updates").Array().Index(0).Value().Document().Lookup("u", "$set", "exported").Boolean(); !got {
			mt.Fatal("expected exported=true in update")
		}
	})
}

func TestListReviewNotesUsesPageAndSize(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("counts total and applies pagination", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		noteID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.review_notes", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(22)}}),
			mtest.CreateCursorResponse(0, "test.review_notes", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: noteID},
				{Key: "content", Value: "第 2 页"},
				{Key: "exported", Value: false},
			}),
		)

		result, err := store.ListReviewNotes(context.Background(), ReviewNoteFilter{Page: 2, PageSize: 10})
		if err != nil {
			mt.Fatalf("ListReviewNotes error: %v", err)
		}
		if result.Total != 22 || result.Page != 2 || result.PageSize != 10 || len(result.Items) != 1 {
			mt.Fatalf("result = %#v", result)
		}

		count := mt.GetStartedEvent()
		find := mt.GetStartedEvent()
		if count.CommandName != "aggregate" || count.Command.Lookup("aggregate").StringValue() != "review_notes" {
			mt.Fatalf("first command = %s %s, want count review_notes", count.CommandName, count.Command.Lookup("aggregate").StringValue())
		}
		if got := find.Command.Lookup("skip").Int64(); got != 10 {
			mt.Fatalf("skip = %d, want 10", got)
		}
		if got := find.Command.Lookup("limit").Int64(); got != 10 {
			mt.Fatalf("limit = %d, want 10", got)
		}
	})
}

func TestClaimSelectedReviewNotesCreatesBatchAndKeepsFiveExports(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("exports selected notes then prunes older export batches", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		actorID := primitive.NewObjectID()
		noteID := primitive.NewObjectID()
		oldBatchID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.review_notes", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: noteID},
				{Key: "actorId", Value: actorID},
				{Key: "content", Value: "选中记录"},
				{Key: "exported", Value: false},
			}),
			mtest.CreateSuccessResponse(bson.E{Key: "insertedId", Value: primitive.NewObjectID()}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}, bson.E{Key: "nModified", Value: int32(1)}),
			mtest.CreateCursorResponse(0, "test.review_note_exports", mtest.FirstBatch, bson.D{{Key: "_id", Value: oldBatchID}}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}),
		)

		notes, batch, err := store.ClaimSelectedReviewNotes(context.Background(), actorID, []primitive.ObjectID{noteID}, "复盘笔记.md")
		if err != nil {
			mt.Fatalf("ClaimSelectedReviewNotes error: %v", err)
		}
		if len(notes) != 1 || notes[0].ID != noteID || batch.NoteCount != 1 {
			mt.Fatalf("notes=%#v batch=%#v", notes, batch)
		}
		if !strings.Contains(batch.Content, "选中记录") {
			mt.Fatalf("expected export batch to store markdown content snapshot, got %q", batch.Content)
		}

		findNotes := mt.GetStartedEvent()
		insertBatch := mt.GetStartedEvent()
		updateNotes := mt.GetStartedEvent()
		findOldBatches := mt.GetStartedEvent()
		deleteOldBatches := mt.GetStartedEvent()
		if !strings.Contains(findNotes.Command.String(), noteID.Hex()) {
			mt.Fatalf("find selected notes should include note id, got %s", findNotes.Command.String())
		}
		if insertBatch.Command.Lookup("insert").StringValue() != "review_note_exports" {
			mt.Fatalf("insert collection = %s", insertBatch.Command.Lookup("insert").StringValue())
		}
		if updateNotes.Command.Lookup("update").StringValue() != "review_notes" {
			mt.Fatalf("update collection = %s", updateNotes.Command.Lookup("update").StringValue())
		}
		if findOldBatches.Command.Lookup("find").StringValue() != "review_note_exports" {
			mt.Fatalf("prune find collection = %s", findOldBatches.Command.Lookup("find").StringValue())
		}
		if deleteOldBatches.Command.Lookup("delete").StringValue() != "review_note_exports" {
			mt.Fatalf("prune delete collection = %s", deleteOldBatches.Command.Lookup("delete").StringValue())
		}
	})
}

func TestDeleteReviewNoteDeletesSingleRecord(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("deletes a selected review note", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		noteID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: int32(1)}))

		if err := store.DeleteReviewNote(context.Background(), noteID); err != nil {
			mt.Fatalf("DeleteReviewNote error: %v", err)
		}

		event := mt.GetStartedEvent()
		if event.CommandName != "delete" || event.Command.Lookup("delete").StringValue() != "review_notes" {
			mt.Fatalf("command = %s %s, want delete review_notes", event.CommandName, event.Command.Lookup("delete").StringValue())
		}
		if got := event.Command.Lookup("deletes").Array().Index(0).Value().Document().Lookup("q", "_id").ObjectID(); got != noteID {
			mt.Fatalf("deleted id = %v, want %v", got, noteID)
		}
	})
}

func TestReviewNoteCountsUsesTotalAndUnexportedQueries(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("counts total and unexported notes", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.review_notes", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(20)}}),
			mtest.CreateCursorResponse(0, "test.review_notes", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(15)}}),
		)

		counts, err := store.ReviewNoteCounts(context.Background())
		if err != nil {
			mt.Fatalf("ReviewNoteCounts error: %v", err)
		}
		if counts.Total != 20 || counts.Unexported != 15 {
			mt.Fatalf("counts = %#v, want total 20 unexported 15", counts)
		}

		total := mt.GetStartedEvent()
		unexported := mt.GetStartedEvent()
		if total.CommandName != "aggregate" || unexported.CommandName != "aggregate" {
			mt.Fatalf("commands = %s, %s; want aggregate/aggregate", total.CommandName, unexported.CommandName)
		}
		if total.Command.Lookup("aggregate").StringValue() != "review_notes" || unexported.Command.Lookup("aggregate").StringValue() != "review_notes" {
			mt.Fatalf("count collections = %s / %s", total.Command.Lookup("aggregate").StringValue(), unexported.Command.Lookup("aggregate").StringValue())
		}
		if !strings.Contains(unexported.Command.String(), "exported") || !strings.Contains(unexported.Command.String(), "false") {
			mt.Fatalf("unexported command should include exported=false, got %s", unexported.Command.String())
		}
	})
}

func TestGetReviewNoteExportLoadsBatchNotes(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("loads export batch snapshot without requiring source notes", func(mt *mtest.T) {
		store := &MongoStore{db: mt.DB}
		batchID := primitive.NewObjectID()
		noteID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "test.review_note_exports", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: batchID},
				{Key: "noteIds", Value: bson.A{noteID}},
				{Key: "noteCount", Value: int32(1)},
				{Key: "filename", Value: "复盘笔记.md"},
				{Key: "content", Value: "# 快照内容"},
			}),
		)

		batch, content, err := store.GetReviewNoteExport(context.Background(), batchID)
		if err != nil {
			mt.Fatalf("GetReviewNoteExport error: %v", err)
		}
		if batch.ID != batchID || batch.Filename != "复盘笔记.md" {
			mt.Fatalf("batch = %#v", batch)
		}
		if content != "# 快照内容" {
			mt.Fatalf("content = %q", content)
		}
		findBatch := mt.GetStartedEvent()
		if findBatch.Command.Lookup("find").StringValue() != "review_note_exports" {
			mt.Fatalf("first collection = %s", findBatch.Command.Lookup("find").StringValue())
		}
	})
}

func TestReviewNoteModelHasExportFields(t *testing.T) {
	note := models.ReviewNote{Exported: true, ExportBatchID: primitive.NewObjectID()}
	if !note.Exported || note.ExportBatchID.IsZero() {
		t.Fatalf("review note export fields not available: %#v", note)
	}
}
