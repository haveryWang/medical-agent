# Repository Map

## Root

- `README.md`: minimal project readme.
- `AGENTS.md`: project-level instructions for Codex and compatible agents.
- `.codex/`: Codex-specific project assets.
- `openspec/`: OpenSpec configuration, specs, and change proposals.
- `backend/`: Go backend service.
- `frontend/`: React frontend service.
- `docs/`: architecture, API, runbook, and execution-record documentation.
- `docker-compose.yml`: local MongoDB and Qdrant services.

## Codex

- `.codex/skills/`: project-local skills for OpenSpec workflows.
- `.codex/wiki/`: persistent project notes for Codex sessions.

## OpenSpec

- `openspec/config.yaml`: OpenSpec configuration.
- `openspec/specs/`: accepted specs. Currently empty.
- `openspec/changes/build-rag-knowledge-chat-system/`: active RAG system change proposal and task records.
- `openspec/changes/archive/`: archived completed changes. Currently empty.

## Backend

- `backend/cmd/server/`: server entrypoint.
- `backend/internal/app/`: application wiring.
- `backend/internal/httpapi/`: chi router, middleware, handlers, requests, responses.
- `backend/internal/store/`: Mongo store split by domain.
- `backend/internal/ingestion/`: document ingestion worker.
- `backend/internal/rag/`: retrieval and answer assembly.
- `backend/internal/providers/`: DeepSeek and Qwen adapters.
- `backend/internal/vector/`: Qdrant adapter.

## Frontend

- `frontend/src/main.jsx`: mount entrypoint.
- `frontend/src/App.jsx`: route table.
- `frontend/src/api/`: API client and SSE parsing.
- `frontend/src/contexts/`: auth context.
- `frontend/src/layouts/`: authenticated shell.
- `frontend/src/pages/`: top-level pages.
- `frontend/src/features/`: chat and knowledge modules.
