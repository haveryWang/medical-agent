# Architecture Notes

The RAG knowledge-base chat system has been implemented as a full-stack project.

## Current Known Structure

The repository currently contains:

- OpenSpec configuration and directories
- Project-local Codex skills
- Codex wiki pages
- OpenSpec change `build-rag-knowledge-chat-system`
- Architecture, API, and runbook docs under `docs/`
- Go backend under `backend/`
- React frontend under `frontend/`

## Implemented System

- Frontend: JavaScript + React + React Router.
- Backend: Go + chi router.
- Primary database: MongoDB for all non-vector data.
- Vector database: Qdrant behind an internal vector client.
- Chat model provider: DeepSeek, configured through backend-only environment
  variables.
- Embedding model provider: Qwen3-Embedding, configured separately from
  DeepSeek chat settings.

## Implemented Product Areas

- Login and session management.
- Agent chat workspace with streaming answers.
- Knowledge-base management and document upload.
- RAG ingestion, Qwen3 embedding, retrieval, prompt assembly, citations, and
  answer details.
- Full startup, build, API, architecture, and execution-record docs.

## Notable Boundaries

- `backend/internal/httpapi/`: chi router, middleware, handlers, requests, responses.
- `backend/internal/store/`: MongoDB access split by auth, knowledge, ingestion, chat, RAG, audit, seed.
- `frontend/src/api/`: REST client and SSE parser.
- `frontend/src/contexts/`: authentication state.
- `frontend/src/features/`: chat and knowledge-base business modules.
