package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"medical-agent/backend/internal/models"
	"medical-agent/backend/internal/reviewnotes"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const maxReviewNoteExports = 5

type ReviewNoteCounts struct {
	Total      int64 `json:"total"`
	Unexported int64 `json:"unexported"`
}

type ReviewNoteFilter struct {
	Page     int64
	PageSize int64
}

type ReviewNoteListResult struct {
	Items    []models.ReviewNote `json:"items"`
	Total    int64               `json:"total"`
	Page     int64               `json:"page"`
	PageSize int64               `json:"size"`
}

func (s *MongoStore) CreateReviewNote(ctx context.Context, actorID primitive.ObjectID, content string) (models.ReviewNote, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return models.ReviewNote{}, errors.New("复盘笔记内容不能为空")
	}
	now := time.Now()
	note := models.ReviewNote{
		ID:        primitive.NewObjectID(),
		ActorID:   actorID,
		Content:   content,
		Exported:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := s.db.Collection("review_notes").InsertOne(ctx, note)
	return note, err
}

func (s *MongoStore) ListReviewNotes(ctx context.Context, filter ReviewNoteFilter) (ReviewNoteListResult, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}
	total, err := s.db.Collection("review_notes").CountDocuments(ctx, bson.M{})
	if err != nil {
		return ReviewNoteListResult{}, err
	}
	cursor, err := s.db.Collection("review_notes").Find(ctx, bson.M{}, options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip((page-1)*pageSize).
		SetLimit(pageSize))
	if err != nil {
		return ReviewNoteListResult{}, err
	}
	defer cursor.Close(ctx)
	var notes []models.ReviewNote
	if err := cursor.All(ctx, &notes); err != nil {
		return ReviewNoteListResult{}, err
	}
	return ReviewNoteListResult{Items: notes, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *MongoStore) ReviewNoteCounts(ctx context.Context) (ReviewNoteCounts, error) {
	total, err := s.db.Collection("review_notes").CountDocuments(ctx, bson.M{})
	if err != nil {
		return ReviewNoteCounts{}, err
	}
	unexported, err := s.db.Collection("review_notes").CountDocuments(ctx, bson.M{"exported": false})
	if err != nil {
		return ReviewNoteCounts{}, err
	}
	return ReviewNoteCounts{Total: total, Unexported: unexported}, nil
}

func (s *MongoStore) ClaimUnexportedReviewNotes(ctx context.Context, actorID primitive.ObjectID, filename string) ([]models.ReviewNote, models.ReviewNoteExport, error) {
	cursor, err := s.db.Collection("review_notes").Find(ctx, bson.M{"exported": false}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	defer cursor.Close(ctx)
	var notes []models.ReviewNote
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	if len(notes) == 0 {
		return nil, models.ReviewNoteExport{}, errors.New("没有未导出的复盘笔记")
	}

	now := time.Now()
	noteIDs := make([]primitive.ObjectID, 0, len(notes))
	for _, note := range notes {
		noteIDs = append(noteIDs, note.ID)
	}
	batch := models.ReviewNoteExport{
		ID:        primitive.NewObjectID(),
		ActorID:   actorID,
		NoteIDs:   noteIDs,
		NoteCount: len(noteIDs),
		Filename:  filename,
		Content:   reviewnotes.RenderMarkdown(notes, now),
		CreatedAt: now,
	}
	if _, err := s.db.Collection("review_note_exports").InsertOne(ctx, batch); err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	_, err = s.db.Collection("review_notes").UpdateMany(ctx, bson.M{"_id": bson.M{"$in": noteIDs}, "exported": false}, bson.M{"$set": bson.M{
		"exported":      true,
		"exportBatchId": batch.ID,
		"exportedAt":    now,
		"updatedAt":     now,
	}})
	if err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	return notes, batch, nil
}

func (s *MongoStore) ClaimSelectedReviewNotes(ctx context.Context, actorID primitive.ObjectID, noteIDs []primitive.ObjectID, filename string) ([]models.ReviewNote, models.ReviewNoteExport, error) {
	if len(noteIDs) == 0 {
		return nil, models.ReviewNoteExport{}, errors.New("请选择要导出的复盘笔记")
	}
	cursor, err := s.db.Collection("review_notes").Find(ctx, bson.M{"_id": bson.M{"$in": noteIDs}}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	defer cursor.Close(ctx)
	var notes []models.ReviewNote
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	if len(notes) == 0 {
		return nil, models.ReviewNoteExport{}, errors.New("请选择要导出的复盘笔记")
	}

	now := time.Now()
	exportedIDs := make([]primitive.ObjectID, 0, len(notes))
	for _, note := range notes {
		exportedIDs = append(exportedIDs, note.ID)
	}
	batch := models.ReviewNoteExport{
		ID:        primitive.NewObjectID(),
		ActorID:   actorID,
		NoteIDs:   exportedIDs,
		NoteCount: len(exportedIDs),
		Filename:  filename,
		Content:   reviewnotes.RenderMarkdown(notes, now),
		CreatedAt: now,
	}
	if _, err := s.db.Collection("review_note_exports").InsertOne(ctx, batch); err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	_, err = s.db.Collection("review_notes").UpdateMany(ctx, bson.M{"_id": bson.M{"$in": exportedIDs}}, bson.M{"$set": bson.M{
		"exported":      true,
		"exportBatchId": batch.ID,
		"exportedAt":    now,
		"updatedAt":     now,
	}})
	if err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	if err := s.pruneReviewNoteExports(ctx, maxReviewNoteExports); err != nil {
		return nil, models.ReviewNoteExport{}, err
	}
	return notes, batch, nil
}

func (s *MongoStore) DeleteReviewNote(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.db.Collection("review_notes").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("复盘笔记不存在")
	}
	return nil
}

func (s *MongoStore) ListReviewNoteExports(ctx context.Context, limit int64) ([]models.ReviewNoteExport, error) {
	if limit <= 0 || limit > maxReviewNoteExports {
		limit = maxReviewNoteExports
	}
	cursor, err := s.db.Collection("review_note_exports").Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var exports []models.ReviewNoteExport
	if err := cursor.All(ctx, &exports); err != nil {
		return nil, err
	}
	return exports, nil
}

func (s *MongoStore) pruneReviewNoteExports(ctx context.Context, keep int64) error {
	if keep <= 0 {
		keep = maxReviewNoteExports
	}
	cursor, err := s.db.Collection("review_note_exports").Find(ctx, bson.M{}, options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(keep).
		SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	var old []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := cursor.All(ctx, &old); err != nil {
		return err
	}
	if len(old) == 0 {
		return nil
	}
	ids := make([]primitive.ObjectID, 0, len(old))
	for _, item := range old {
		ids = append(ids, item.ID)
	}
	_, err = s.db.Collection("review_note_exports").DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
	return err
}

func (s *MongoStore) GetReviewNoteExport(ctx context.Context, id primitive.ObjectID) (models.ReviewNoteExport, string, error) {
	var batch models.ReviewNoteExport
	if err := s.db.Collection("review_note_exports").FindOne(ctx, bson.M{"_id": id}).Decode(&batch); err != nil {
		return models.ReviewNoteExport{}, "", err
	}
	if batch.Content != "" {
		return batch, batch.Content, nil
	}
	cursor, err := s.db.Collection("review_notes").Find(ctx, bson.M{"_id": bson.M{"$in": batch.NoteIDs}}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return models.ReviewNoteExport{}, "", err
	}
	defer cursor.Close(ctx)
	var notes []models.ReviewNote
	if err := cursor.All(ctx, &notes); err != nil {
		return models.ReviewNoteExport{}, "", err
	}
	return batch, reviewnotes.RenderMarkdown(notes, batch.CreatedAt), nil
}
