package ingestion

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"regexp"
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

type chunkPart struct {
	Section string
	Text    string
}

var headingLineRe = regexp.MustCompile(`^(?:\s{0,3}(?:#{1,6}\s+|[\p{Han}A-Za-z0-9]{1,30}[:：]\s*))(.+?)\s*$`)

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
	text, err := extractText(doc)
	if err != nil {
		_ = w.store.FailJob(ctx, job, err.Error())
		return
	}
	parts := chunkDocument(text, 200)
	vectors, err := w.qwen.Embed(ctx, chunkTexts(parts))
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
			Text:            part.Text,
			Section:         part.Section,
			ChunkIndex:      i,
			VectorID:        vectorID,
			Checksum:        checksum(part.Section + "\n" + part.Text),
			CreatedAt:       now,
		}
		chunks = append(chunks, chunk)
		points = append(points, vector.Point{
			ID:     vectorID,
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

func chunkDocument(text string, maxLen int) []chunkPart {
	text = normalizeChunkSource(text)
	if text == "" {
		return []chunkPart{{Section: "空文档", Text: "空文档"}}
	}

	lines := strings.Split(text, "\n")
	sections := make([]chunkPart, 0)
	currentSection := "正文"
	var paragraph strings.Builder

	flushParagraph := func() {
		content := strings.TrimSpace(paragraph.String())
		paragraph.Reset()
		if content == "" {
			return
		}
		sections = append(sections, splitParagraph(currentSection, content, maxLen)...)
	}

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			flushParagraph()
			continue
		}
		if heading := detectHeading(line); heading != "" {
			flushParagraph()
			currentSection = heading
			continue
		}
		if paragraph.Len() > 0 {
			paragraph.WriteByte(' ')
		}
		paragraph.WriteString(line)
	}
	flushParagraph()

	if len(sections) == 0 {
		return []chunkPart{{Section: currentSection, Text: text}}
	}
	return sections
}

func chunkTexts(parts []chunkPart) []string {
	texts := make([]string, 0, len(parts))
	for _, part := range parts {
		texts = append(texts, part.Text)
	}
	return texts
}

func splitParagraph(section, text string, maxLen int) []chunkPart {
	if maxLen <= 0 {
		maxLen = 200
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return nil
	}
	if len(runes) <= maxLen {
		return []chunkPart{{Section: section, Text: string(runes)}}
	}
	parts := make([]chunkPart, 0, (len(runes)/maxLen)+1)
	for start := 0; start < len(runes); {
		end := start + maxLen
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, chunkPart{Section: section, Text: string(runes[start:end])})
		start = end
	}
	return parts
}

func detectHeading(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	if strings.HasPrefix(line, "#") {
		line = strings.TrimLeft(line, "#")
		return strings.TrimSpace(line)
	}
	if matches := headingLineRe.FindStringSubmatch(line); len(matches) == 2 {
		if len([]rune(line)) <= 48 {
			return line
		}
	}
	if looksLikeStandaloneHeading(line) {
		return line
	}
	return ""
}

func looksLikeStandaloneHeading(line string) bool {
	if len([]rune(line)) > 32 || endsWithSentencePunctuation(line) {
		return false
	}
	if strings.ContainsAny(line, " \t") {
		return false
	}
	return true
}

func endsWithSentencePunctuation(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	last := []rune(line)[len([]rune(line))-1]
	return strings.ContainsRune("。！？!?；;，,、.", last)
}

func normalizeChunkSource(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSpace(text)
	return text
}

func checksum(text string) string {
	sum := sha1.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}
