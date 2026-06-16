package httpapi

import (
	"log"

	"medical-agent/backend/internal/config"
	"medical-agent/backend/internal/ingestion"
	"medical-agent/backend/internal/rag"
	"medical-agent/backend/internal/store"
	"medical-agent/backend/internal/vector"
)

type API struct {
	cfg    config.Config
	store  *store.MongoStore
	rag    *rag.Service
	worker *ingestion.Worker
	vector *vector.Client
	logger *log.Logger
}

func New(cfg config.Config, store *store.MongoStore, ragService *rag.Service, worker *ingestion.Worker, vectorClient *vector.Client, logger *log.Logger) *API {
	return &API{cfg: cfg, store: store, rag: ragService, worker: worker, vector: vectorClient, logger: logger}
}
