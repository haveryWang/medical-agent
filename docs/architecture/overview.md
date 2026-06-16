# RAG Knowledge-Base Chat System Architecture

## Purpose

This document records the implemented architecture for the medical knowledge-base conversation system.

## Product Areas

- Login and session management.
- Agent conversation workspace with streaming answers.
- Knowledge-base management with list, filters, document upload, and indexing status.
- RAG ingestion and retrieval pipeline.
- Documentation and task execution records.

## Target Stack

- Frontend: JavaScript + React.
- Frontend framework libraries: React Router for page routing, React Context and hooks for session and workspace state.
- Backend: Go with chi router and chi middleware.
- Source-of-truth database: MongoDB.
- Vector database: Qdrant by default, behind a vector-store interface.
- Model providers: DeepSeek for chat generation and Qwen3-Embedding for vector generation, configured from MongoDB model settings with environment variables as bootstrap fallback.

## Module Map

```text
frontend/
  src/
    api/        REST client and SSE stream parser
    contexts/   Auth context, token lifecycle, current user
    layouts/    Top-level authenticated shell
    pages/      Login, chat, and knowledge-base pages
    features/   Chat and knowledge-base feature modules
    components/ Shared presentational components
    utils/      Formatting helpers

backend/
  cmd/server/       HTTP server entrypoint
  internal/app/     Dependency wiring, worker lifecycle, router creation
  internal/httpapi/ chi router, middleware, request DTOs, handlers, response helpers
  internal/store/   MongoDB store split by auth, chat, knowledge, ingestion, RAG, audit
  internal/ingestion/
  internal/rag/
  internal/vector/
  internal/providers/deepseek/
  internal/providers/qwen/
```

## Storage Responsibilities

MongoDB stores users, sessions, roles, conversations, messages, knowledge bases, document metadata, chunks, ingestion jobs, audit logs, and model configuration.

Qdrant stores embedding vectors and searchable payload fields. MongoDB remains canonical for source text, metadata, permissions, and audit records.

Secrets such as `DEEPSEEK_API_KEY` and `QWEN_EMBEDDING_API_KEY` must not be stored in source code or returned to the frontend. Operators can save provider API keys and model names through the system settings modal; backend ingestion and chat execution reads the effective values from MongoDB and falls back to environment variables when no database value exists.

## Design Source

The initial UI and product scope come from `design.png` and the OpenSpec change at `openspec/changes/build-rag-knowledge-chat-system/`.

## Implemented Files

- `backend/cmd/server/main.go`: Go backend entrypoint.
- `backend/internal/httpapi/router.go`: chi router, CORS, route groups, protected routes.
- `backend/internal/httpapi/*_handlers.go`: REST and SSE API handlers split by domain.
- `backend/internal/httpapi/middleware.go`: request ID, auth, permission middleware.
- `backend/internal/store/mongo.go`: MongoDB connection, indexes, shared store types.
- `backend/internal/store/*.go`: Mongo repositories split by auth, chat, knowledge, ingestion, RAG, audit, seed data.
- `backend/internal/providers/deepseek/deepseek.go`: DeepSeek chat streaming adapter.
- `backend/internal/providers/qwen/qwen.go`: Qwen3-Embedding adapter and local deterministic demo fallback.
- `backend/internal/vector/qdrant.go`: Qdrant collection, upsert, and search adapter.
- `backend/internal/ingestion/ingestion.go`: upload ingestion worker, text chunking, embedding, vector indexing.
- `frontend/src/main.jsx`: React mount point.
- `frontend/src/App.jsx`: React Router route table and protected routes.
- `frontend/src/api/client.js`: API client and streaming conversation request.
- `frontend/src/contexts/AuthContext.jsx`: session state and auth API integration.
- `frontend/src/features/chat/`: conversation list, messages, composer, answer detail, streaming hook.
- `frontend/src/features/knowledge/`: filters, table, upload panel, workspace hook.
- `frontend/src/styles.css`: design-image-matched UI styles.
- `backend/scripts/mongo-init.js`: MongoDB collection and index initialization.
- `backend/scripts/qdrant-init.sh`: Qdrant collection initialization.
