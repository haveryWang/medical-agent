package ingestion

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"medical-agent/backend/internal/models"
	"medical-agent/backend/internal/providers/qwen"
	"medical-agent/backend/internal/store"
	"medical-agent/backend/internal/vector"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Worker struct {
	store  *store.MongoStore
	qwen   *qwen.Client
	vector *vector.Client
}

func NewWorker(store *store.MongoStore, qwen *qwen.Client, vector *vector.Client) *Worker {
	return &Worker{store: store, qwen: qwen, vector: vector}
}

func (w *Worker) RunOnce(ctx context.Context) {
	jobs, err := w.store.PendingJobs(ctx, 5)
	if err != nil {
		return
	}
	for _, job := range jobs {
		w.process(ctx, job)
	}
}

func (w *Worker) process(ctx context.Context, job models.IngestionJob) {
	doc, err := w.store.GetDocument(ctx, job.DocumentID)
	if err != nil {
		_ = w.store.FailJob(ctx, job, err.Error())
		return
	}
	text, err := extractText(doc.StoragePath)
	if err != nil {
		_ = w.store.FailJob(ctx, job, err.Error())
		return
	}
	parts := chunkText(text, 650, 90)
	vectors, err := w.qwen.Embed(ctx, parts)
	if err != nil {
		_ = w.store.FailJob(ctx, job, err.Error())
		return
	}
	now := time.Now()
	chunks := make([]models.Chunk, 0, len(parts))
	points := make([]vector.Point, 0, len(parts))
	for i, part := range parts {
		id := primitive.NewObjectID()
		vectorID := vector.PointIDFromObjectID(id)
		chunk := models.Chunk{
			ID:              id,
			KnowledgeBaseID: job.KnowledgeBaseID,
			DocumentID:      job.DocumentID,
			Text:            part,
			Section:         doc.FileName,
			ChunkIndex:      i,
			VectorID:        vectorID,
			Checksum:        checksum(part),
			CreatedAt:       now,
		}
		chunks = append(chunks, chunk)
		points = append(points, vector.Point{
			ID:      vectorID,
			Vector: vectors[i],
			Payload: map[string]any{
				"knowledgeBaseId": job.KnowledgeBaseID.Hex(),
				"documentId":      job.DocumentID.Hex(),
				"chunkId":         id.Hex(),
			},
		})
	}
	if err := w.vector.EnsureCollection(ctx); err != nil {
		_ = w.store.FailJob(ctx, job, err.Error())
		return
	}
	if err := w.vector.Upsert(ctx, points); err != nil {
		_ = w.store.FailJob(ctx, job, err.Error())
		return
	}
	if err := w.store.CompleteJobWithChunks(ctx, job, chunks); err != nil {
		_ = w.store.FailJob(ctx, job, err.Error())
	}
}

func extractText(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 5*1024*1024))
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		text = "文档 " + filepath.Base(path) + " 已上传。当前演示解析器保留原始文件元数据，正式环境可接入 PDF/Word/Excel 解析器提取正文。"
	}
	return text, nil
}

func chunkText(text string, size int, overlap int) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return []string{"空文档"}
	}
	if len(runes) <= size {
		return []string{string(runes)}
	}
	var chunks []string
	for start := 0; start < len(runes); {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
		if end == len(runes) {
			break
		}
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}

func checksum(text string) string {
	sum := sha1.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}
