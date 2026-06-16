package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/httpapi"
	"medical-agent/backend/internal/ingestion"
	"medical-agent/backend/internal/providers/deepseek"
	"medical-agent/backend/internal/providers/qwen"
	"medical-agent/backend/internal/rag"
	"medical-agent/backend/internal/store"
	"medical-agent/backend/internal/vector"
)

type App struct {
	cancel context.CancelFunc
	store  *store.MongoStore
	router http.Handler
}

func New(ctx context.Context, cfg config.Config, logger *log.Logger) (*App, error) {
	mongoStore, err := store.NewMongoStore(ctx, cfg)
	if err != nil {
		return nil, err
	}
	qwenClient := qwen.New(cfg)
	vectorClient := vector.New(cfg)
	deepSeekClient := deepseek.New(cfg)
	if err := vectorClient.EnsureCollection(ctx); err != nil {
		logger.Printf("Qdrant collection 初始化暂未完成: %v", err)
	}
	worker := ingestion.NewWorker(mongoStore, qwenClient, vectorClient)
	workerCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				worker.RunOnce(workerCtx)
			}
		}
	}()
	ragService := rag.New(cfg, mongoStore, qwenClient, vectorClient, deepSeekClient)
	api := httpapi.New(cfg, mongoStore, ragService, worker, logger)
	return &App{cancel: cancel, store: mongoStore, router: api.Router()}, nil
}

func (a *App) Router() http.Handler {
	return a.router
}

func (a *App) Close(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
	_ = a.store.Close(ctx)
}
