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
- Model provider: Volcengine Ark-compatible API settings are stored in MongoDB
  and exposed through the backend settings API.
- Chat model: `deepseek-v4-flash-260425`.
- Embedding model: `doubao-embedding-vision-251215` with 2048-dimensional vectors.

## Implemented Product Areas

- Login and session management.
- Agent chat workspace with streaming answers.
- Knowledge-base management and document upload.
- RAG ingestion, document preprocessing, embedding, retrieval, prompt assembly,
  citations, and answer details.
- Document preprocessing supports PDF, Word `.docx`, Excel `.xlsx/.xls`,
  Markdown, TXT, and CSV. Legacy Word `.doc` is rejected with a conversion
  message.
- Knowledge-base documents can be downloaded from MongoDB raw content, previewed
  as preprocessed text, and inspected by stored chunk text.
- Full startup, build, API, architecture, and execution-record docs.

## Notable Boundaries

- `backend/internal/httpapi/`: chi router, middleware, handlers, requests, responses.
- `backend/internal/store/`: MongoDB access split by auth, knowledge, ingestion, chat, RAG, audit, seed.
- `frontend/src/api/`: REST client and SSE parser.
- `frontend/src/contexts/`: authentication state.
- `frontend/src/features/`: chat and knowledge-base business modules.
