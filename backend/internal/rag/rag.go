package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/models"
	"medical-agent/backend/internal/providers/deepseek"
	"medical-agent/backend/internal/providers/qwen"
	"medical-agent/backend/internal/store"
	"medical-agent/backend/internal/vector"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Service struct {
	cfg      config.Config
	store    *store.MongoStore
	qwen     *qwen.Client
	vector   *vector.Client
	deepseek *deepseek.Client
}

type Retrieval struct {
	Citations []models.Citation
	Context   string
}

func New(cfg config.Config, store *store.MongoStore, qwen *qwen.Client, vector *vector.Client, deepseek *deepseek.Client) *Service {
	return &Service{cfg: cfg, store: store, qwen: qwen, vector: vector, deepseek: deepseek}
}

func (s *Service) Retrieve(ctx context.Context, question string, kbIDs []primitive.ObjectID) (Retrieval, error) {
	if len(kbIDs) == 0 {
		return Retrieval{}, nil
	}
	queryVectors, err := s.qwen.Embed(ctx, []string{question})
	if err != nil {
		return Retrieval{}, err
	}
	kbHex := make([]string, 0, len(kbIDs))
	for _, id := range kbIDs {
		kbHex = append(kbHex, id.Hex())
	}
	results, err := s.vector.Search(ctx, queryVectors[0], kbHex, s.cfg.RetrievalTopK)
	if err != nil || len(results) == 0 {
		chunks, fallbackErr := s.store.SearchChunks(ctx, kbIDs, int64(s.cfg.RetrievalTopK))
		if fallbackErr != nil {
			if err != nil {
				return Retrieval{}, err
			}
			return Retrieval{}, fallbackErr
		}
		return fromChunks(chunks, nil), nil
	}
	vectorIDs := make([]string, 0, len(results))
	scores := map[string]float64{}
	for _, result := range results {
		vectorIDs = append(vectorIDs, result.ID)
		scores[result.ID] = result.Score
	}
	chunks, err := s.store.GetChunksByVectorIDs(ctx, vectorIDs)
	if err != nil {
		return Retrieval{}, err
	}
	return fromChunks(chunks, scores), nil
}

func (s *Service) StreamAnswer(ctx context.Context, question string, retrieval Retrieval, onDelta func(string) error) (string, time.Duration, error) {
	start := time.Now()
	system := "你是医院行政智策平台的智能问答助手。"
	if strings.TrimSpace(retrieval.Context) != "" {
		system += "必须优先根据提供的知识库上下文回答；如果上下文不足，明确说明无法从知识库确认，不要编造来源。\n\n知识库上下文：\n" + retrieval.Context
	}
	var builder strings.Builder
	err := s.deepseek.StreamChat(ctx, []deepseek.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: question},
	}, func(delta string) error {
		builder.WriteString(delta)
		return onDelta(delta)
	})
	return builder.String(), time.Since(start), err
}

func fromChunks(chunks []models.Chunk, scores map[string]float64) Retrieval {
	var contextBuilder strings.Builder
	citations := make([]models.Citation, 0, len(chunks))
	for i, chunk := range chunks {
		score := 0.66
		if scores != nil {
			if value, ok := scores[chunk.VectorID]; ok {
				score = value
			}
		}
		title := chunk.Section
		if title == "" {
			title = "知识片段"
		}
		snippet := chunk.Text
		if len([]rune(snippet)) > 160 {
			snippet = string([]rune(snippet)[:160]) + "..."
		}
		contextBuilder.WriteString(fmt.Sprintf("[%d] %s\n%s\n\n", i+1, title, chunk.Text))
		citations = append(citations, models.Citation{
			ChunkID:         chunk.ID,
			DocumentID:      chunk.DocumentID,
			KnowledgeBaseID: chunk.KnowledgeBaseID,
			Title:           title,
			Snippet:         snippet,
			Score:           score,
		})
	}
	if contextBuilder.Len() == 0 {
		contextBuilder.WriteString("未检索到高置信度知识库上下文。")
	}
	return Retrieval{Citations: citations, Context: contextBuilder.String()}
}
