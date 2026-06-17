## Context

The repository is currently an early-stage project with OpenSpec scaffolding, Codex wiki notes, and a design image at `design.png`. The requested system is a full-stack medical knowledge-base management and RAG chat platform.

The design image describes:

- Login page branded as "医院行政智策平台".
- Main chat workspace with conversation list, new conversation action, search, message area, streaming-style answer display, message actions, answer detail drawer, cited knowledge references, retrieval strategy, and prompt context preview.
- Knowledge-base management page with sidebar navigation, filters, list table, upload action, document upload panel, supported file types, and uploaded file statuses.
- System requirements for RAG-based intelligent Q&A, user roles, chat management, knowledge-base maintenance, vector retrieval, DeepSeek integration, security, and file format support.

Primary technical constraints:

- Frontend: JavaScript + React.
- Backend: Go.
- Vector database: lightweight vector database.
- Non-vector persistence: MongoDB.
- Model providers: DeepSeek for chat generation and Qwen3-Embedding for vector generation, with credentials left as deployment configuration.
- Chat responses: streaming output.
- Documentation: all steps and implementation records must be retained.

## Goals / Non-Goals

**Goals:**

- Define an implementation-ready plan for a complete frontend and backend RAG knowledge-base conversation system.
- Use MongoDB as the source of truth for non-vector application data.
- Use a lightweight vector database for embedding similarity search.
- Provide DeepSeek chat integration and Qwen3-Embedding integration points with separate configuration placeholders for credentials, endpoints, model names, and vector dimensions.
- Support document upload, parsing, chunking, embedding, vector indexing, retrieval, prompt assembly, streaming chat, citations, and answer details.
- Provide explicit documentation requirements for startup, build, API integration, architecture, and task records.
- Plan subagent responsibilities so work can be split by domain and reviewed systematically.

**Non-Goals:**

- Implement application code during explore mode.
- Hardcode production secrets, including DeepSeek API keys.
- Build multi-tenant billing, hospital SSO, or external EMR integrations in the first implementation.
- Guarantee clinical correctness of generated answers beyond grounded retrieval, citations, and clear source boundaries.
- Store all file binaries in MongoDB by default; raw file storage can be local development storage initially and later replaced by object storage.

## Decisions

### Decision: React frontend with feature-oriented modules

Use a React application under `frontend/` organized by feature areas:

- `auth`: login, session storage, route protection.
- `chat`: conversation list, message stream, answer detail drawer, citation display.
- `knowledge`: knowledge-base list, filters, upload panel, ingestion status.
- `shared`: API client, layout, UI primitives, formatting, error handling.

Alternatives considered:

- Server-rendered templates: simpler backend but weak fit for streaming chat and management UI interactions.
- TypeScript: safer long-term, but the explicit requirement is JavaScript + React. The plan can still use JSDoc and runtime validation where useful.

### Decision: Go backend with layered services

Use a Go backend under `backend/` with clear layers:

- HTTP handlers for REST and streaming APIs.
- Application services for auth, chat, knowledge-base management, ingestion, retrieval, and model orchestration.
- Repositories for MongoDB persistence.
- Provider adapters for DeepSeek chat, Qwen3 embeddings, vector database, and document parsing.

Alternatives considered:

- Single large handler package: faster to scaffold but harder to test and maintain.
- Microservices: premature for the initial version and adds operational burden.

### Decision: MongoDB for source-of-truth application data

MongoDB stores users, sessions, conversations, messages, knowledge bases, documents, chunks, ingestion jobs, audit logs, and model configuration metadata.

The vector database stores embeddings and vector payload metadata needed for similarity search. MongoDB keeps canonical records and vector ids.

Alternatives considered:

- Storing vectors in MongoDB only: reduces dependencies but weakens vector search ergonomics and tuning.
- Relational database: stronger schema enforcement, but the requirement specifies MongoDB.

### Decision: Qdrant as the default lightweight vector database

Use Qdrant as the default vector database because it is lightweight for local development, has Docker support, supports payload filtering, and has Go client support. Keep vector access behind an internal interface so another lightweight vector store can replace it later.

Alternatives considered:

- Chroma: common in Python RAG stacks, less natural for a Go-first backend.
- Milvus: powerful but heavier operationally for this project.
- Embedded in-process vector index: simplest deployment but weaker metadata filtering and persistence semantics.

### Decision: DeepSeek for chat and Qwen3-Embedding for embeddings

Use DeepSeek for chat completion through a backend-only provider adapter. The DeepSeek API key is read from environment or deployment configuration, for example `DEEPSEEK_API_KEY`. The default planned chat model is `deepseek-v4-flash-260425`, with `deepseek-v4-pro` available as a configurable higher-reasoning option.

RAG embedding generation uses Qwen3-Embedding through a separate backend-only provider adapter. The Qwen embedding endpoint, API key, model name, and vector dimension are configured separately from DeepSeek chat settings so the embedding service can be operated and tuned independently.

Alternatives considered:

- Use DeepSeek for embeddings: rejected because `deepseek-embed` is not the desired embedding model for this project.
- Use local embeddings by default: avoids external dependency, but increases setup complexity and makes deployment less consistent with the requested Qwen3-Embedding service.

### Decision: SSE for chat streaming

Use Server-Sent Events for the initial streaming endpoint:

- Frontend sends a chat request and consumes `text/event-stream`.
- Backend streams event types such as `message.started`, `message.delta`, `retrieval.sources`, `message.completed`, and `message.error`.

Alternatives considered:

- WebSocket: useful for bidirectional real-time features, but SSE is simpler for one-way answer streaming.
- Chunked JSON without event framing: less explicit for frontend state handling and error recovery.

### Decision: Asynchronous ingestion jobs

Document upload creates an ingestion job. The backend processes parsing, chunking, embedding, and vector indexing asynchronously. The frontend polls or refreshes ingestion status.

Alternatives considered:

- Synchronous ingestion during upload: simpler but poor UX for larger documents.
- External queue service: useful later, but initial implementation can use a Go worker with MongoDB-backed job state.

### Decision: Documentation as a first-class deliverable

Maintain documentation in these locations:

- `docs/architecture/`: architecture, RAG pipeline, data model, security.
- `docs/api/`: REST and streaming API contracts.
- `docs/runbooks/`: frontend/backend startup, build, and operations.
- `docs/records/`: task execution records and subagent handoffs.
- `docs/superpowers/plans/`: implementation plan.
- `.codex/wiki/`: stable project notes for future Codex sessions.

## Architecture Sketch

```text
┌─────────────────────────────┐
│ React Frontend              │
│                             │
│ Login │ Chat │ Knowledge UI │
└──────────────┬──────────────┘
               │ HTTPS / SSE
               ▼
┌─────────────────────────────┐
│ Go Backend API              │
│                             │
│ Auth Handlers               │
│ Chat Stream Handlers        │
│ Knowledge Base Handlers     │
│ Upload / Ingestion Handlers │
└──────────────┬──────────────┘
               │
       ┌───────┼────────────────┬────────────────┐
       ▼       ▼                ▼                ▼
┌─────────┐ ┌──────────┐ ┌──────────────┐ ┌──────────────┐
│ MongoDB │ │ Qdrant   │ │ DeepSeek Chat │ │ Qwen3        │
│ Data    │ │ Vectors  │ │ Provider      │ │ Embeddings   │
└─────────┘ └──────────┘ └──────────────┘ └──────────────┘
```

## RAG Data Flow

```text
Document Upload
   │
   ▼
Store document metadata in MongoDB
   │
   ▼
Create ingestion job
   │
   ▼
Parse document text
   │
   ▼
Chunk text + attach metadata
   │
   ▼
Generate embeddings with Qwen3-Embedding
   │
   ▼
Upsert vectors into Qdrant
   │
   ▼
Store chunk records and vector ids in MongoDB
```

```text
User Question
   │
   ▼
Persist user message
   │
   ▼
Embed query with Qwen3-Embedding
   │
   ▼
Search Qdrant with knowledge-base filters
   │
   ▼
Load chunk metadata/source snippets from MongoDB
   │
   ▼
Assemble grounded prompt
   │
   ▼
Call DeepSeek chat stream
   │
   ▼
Stream answer deltas to frontend
   │
   ▼
Persist final answer, citations, and RAG metadata
```

## Initial Data Model

MongoDB collections:

- `users`: account, password hash, display name, role ids, status.
- `sessions`: user id, token hash, expiry, revoked state.
- `roles`: role name and permissions.
- `knowledge_bases`: name, scenario, tags, department, description, status, retrieval config, counts.
- `documents`: knowledge base id, filename, type, size, storage location, ingestion status, failure reason.
- `chunks`: document id, knowledge base id, text, section metadata, vector id, chunk index, checksum.
- `ingestion_jobs`: document id, status, step, attempts, timestamps, error.
- `conversations`: user id, title, status, selected knowledge-base ids, timestamps.
- `messages`: conversation id, role, content, status, citations, model metadata, timing.
- `audit_logs`: actor, action, target, result, timestamp, request id.
- `model_configs`: DeepSeek base URL, chat model, Qwen embedding endpoint, Qwen embedding model name, vector dimension, and non-secret defaults. Secrets stay in environment or secret manager.

## API Shape

Initial REST endpoints:

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `GET /api/v1/conversations`
- `POST /api/v1/conversations`
- `GET /api/v1/conversations/{id}/messages`
- `POST /api/v1/conversations/{id}/messages:stream`
- `GET /api/v1/messages/{id}/details`
- `GET /api/v1/knowledge-bases`
- `POST /api/v1/knowledge-bases`
- `PATCH /api/v1/knowledge-bases/{id}`
- `POST /api/v1/knowledge-bases/{id}/documents`
- `GET /api/v1/knowledge-bases/{id}/documents`
- `GET /api/v1/ingestion-jobs/{id}`
- `POST /api/v1/ingestion-jobs/{id}:retry`

Streaming endpoint uses SSE with named events and JSON payloads.

## Subagent Workstreams

Use subagents by independent domain, with a controller maintaining the OpenSpec plan and task records:

- Product/spec agent: keeps OpenSpec specs and user-facing requirements aligned with `design.png`.
- UX/frontend agent: implements React layout, routing, state, API client, SSE handling, and frontend docs.
- Backend API agent: implements Go server, auth, API contracts, middleware, and backend docs.
- RAG ingestion agent: implements parsing, chunking, ingestion jobs, Qwen3-Embedding adapter, vector indexing, and related docs.
- RAG chat agent: implements retrieval, prompt assembly, DeepSeek streaming adapter, citations, and answer details.
- Data/storage agent: implements MongoDB repositories, indexes, migrations/seed scripts, Qdrant setup, and data model docs.
- QA/review agent: owns verification matrix, test plans, contract tests, and task record audits.
- Documentation agent: ensures setup, build, API, architecture, and record docs stay current.

Each implementation task should end with:

- Files changed.
- Commands run.
- Test result.
- Documentation updated.
- Remaining risks.

## Risks / Trade-offs

- Qwen3-Embedding availability or dimension mismatch -> Add startup/health checks that verify the configured Qwen3-Embedding model and vector dimension before ingestion runs, and fail ingestion with a clear configuration error if unavailable.
- Medical-domain hallucination risk -> Require grounded prompts, visible citations, no-context fallback behavior, and answer detail inspection.
- Large file ingestion latency -> Use asynchronous jobs and visible build status.
- Vector/Mongo consistency drift -> Store vector ids in MongoDB, use idempotent indexing, and support retry/rebuild operations.
- SSE interruption -> Persist partial message status and allow retry or regenerate actions.
- Secret leakage -> Keep API keys backend-only and out of source code, frontend bundles, logs, and API responses.
- Scope size -> Split execution into independently reviewable milestones and keep documentation records per task.

## Migration Plan

This is a greenfield implementation:

1. Create frontend and backend project skeletons.
2. Add local development configuration for MongoDB and Qdrant.
3. Implement authentication and session baseline.
4. Implement knowledge-base metadata and upload lifecycle.
5. Implement ingestion pipeline and vector indexing.
6. Implement RAG chat streaming and answer details.
7. Complete docs, tests, and verification matrix.

Rollback for early development is file-level or branch-level. Production migration is out of scope until deployment requirements are defined.

## Open Questions

- Which Qwen3-Embedding endpoint, API-compatible protocol, model id, and vector dimensions will be used in the target deployment?
- Should raw uploaded files be stored on local disk for development only, or should object storage be introduced immediately?
- What exact roles are required beyond system administrator, doctor/user, and knowledge-base manager?
- Should conversation access be personal-only, department-scoped, or admin-visible?
- Are there hospital compliance requirements for data retention, audit log retention, and sensitive document handling?
