## 1. Planning And Documentation Baseline

- [x] 1.1 Confirm `design.png` requirements are reflected in OpenSpec specs, design, and task records.
- [x] 1.2 Create `docs/architecture/overview.md` with the system architecture, module map, and storage responsibilities.
- [x] 1.3 Create `docs/architecture/rag-pipeline.md` with ingestion and chat retrieval flows.
- [x] 1.4 Create `docs/api/rest-api.md` with initial REST request and response contracts.
- [x] 1.5 Create `docs/api/streaming-chat.md` with SSE event contracts and frontend handling notes.
- [x] 1.6 Create `docs/runbooks/frontend.md` and `docs/runbooks/backend.md` with startup/build placeholders that implementation tasks must keep current.
- [x] 1.7 Create `docs/records/task-log.md` and require every implementation task to append execution evidence.

## 2. Project Scaffolding

- [x] 2.1 Scaffold `frontend/` as a JavaScript React application with routing, lint/check scripts, environment template, and documented startup command.
- [x] 2.2 Scaffold `backend/` as a Go module with HTTP server entrypoint, configuration loading, structured logging, health endpoint, and documented startup command.
- [x] 2.3 Add local development configuration for MongoDB and Qdrant, including documented connection environment variables.
- [x] 2.4 Add shared API error format, request id propagation, and frontend API client conventions.
- [x] 2.5 Update documentation with actual generated project commands and verification results.

## 3. Authentication And Session Management

- [x] 3.1 Implement MongoDB collections and repository functions for users, roles, permissions, and sessions.
- [x] 3.2 Implement password hashing, login, logout, session validation middleware, and `/api/v1/auth/me`.
- [x] 3.3 Implement frontend login page matching `design.png`, protected routes, session persistence, and logout behavior.
- [x] 3.4 Add tests for successful login, invalid login, missing session, expired session, and role-aware navigation.
- [x] 3.5 Update API and runbook documentation with auth flows and environment requirements.

## 4. Knowledge-Base Management

- [x] 4.1 Implement MongoDB models and repositories for knowledge bases, documents, ingestion jobs, chunks, and audit logs.
- [x] 4.2 Implement backend knowledge-base list, filtering, create, edit, disable, document list, upload, ingestion status, and retry APIs.
- [x] 4.3 Implement frontend sidebar navigation, knowledge-base list, filters, table actions, upload panel, file validation, and ingestion status display matching `design.png`.
- [x] 4.4 Add tests for filters, pagination, file validation, upload success, upload rejection, ingestion status, retry, and audit events.
- [x] 4.5 Update API, architecture, and task-log documentation with knowledge-base behavior and verification evidence.

## 5. RAG Ingestion And Vector Indexing

- [x] 5.1 Implement document parser interfaces and initial parsers for PDF, Word, Excel, Markdown, and text files.
- [x] 5.2 Implement text normalization, chunking, chunk metadata, checksums, and idempotent chunk persistence.
- [x] 5.3 Implement Qwen3-Embedding adapter with separate configuration, vector dimension validation, and startup availability checks.
- [x] 5.4 Implement Qdrant vector store adapter with collection setup, upsert, filtered search, and deletion/exclusion behavior.
- [x] 5.5 Implement asynchronous ingestion worker from upload job to parsed chunks, embeddings, vector upserts, and final status.
- [x] 5.6 Add tests for parser failures, chunking, missing Qwen3 embedding configuration, unavailable Qwen3-Embedding model, vector dimension mismatch, vector upsert, retry, and Mongo/vector consistency.
- [x] 5.7 Update RAG pipeline and backend runbook documentation with actual ingestion commands and troubleshooting notes.

## 6. Agent Chat And DeepSeek Streaming

- [x] 6.1 Implement conversation and message MongoDB repositories with statuses, citations, timing, and retrieval metadata.
- [x] 6.2 Implement conversation list, create conversation, message history, and answer detail APIs.
- [x] 6.3 Implement Qwen3 query embedding, vector retrieval, MongoDB chunk hydration, prompt assembly, and no-context fallback behavior.
- [x] 6.4 Implement DeepSeek chat provider adapter with backend-only `DEEPSEEK_API_KEY` configuration and streaming response handling.
- [x] 6.5 Implement SSE streaming endpoint with `message.started`, `retrieval.sources`, `message.delta`, `message.completed`, and `message.error` events.
- [x] 6.6 Implement frontend chat workspace matching `design.png`, including conversation search, message composer, progressive answer rendering, citations, and answer detail drawer.
- [x] 6.7 Add tests for streaming success, streaming failure, missing DeepSeek key, no-context fallback, citation persistence, and answer detail rendering.
- [x] 6.8 Update streaming API, RAG pipeline, runbook, and task-log documentation with chat verification evidence.

## 7. Security, Observability, And Operations

- [x] 7.1 Add request logging, request ids, audit events, model/retrieval timing, and safe error responses.
- [x] 7.2 Add file upload size/type enforcement, backend access control checks, secret redaction, and frontend forbidden states.
- [x] 7.3 Add MongoDB indexes for users, sessions, conversations, messages, knowledge-base filters, documents, chunks, ingestion jobs, and audit logs.
- [x] 7.4 Add health checks for backend, MongoDB, Qdrant, DeepSeek chat configuration, and Qwen3-Embedding configuration.
- [x] 7.5 Update operations documentation with health checks, common failures, and recovery steps.

## 8. Full-System Verification And Handoff

- [x] 8.1 Create a seed/demo dataset for local development without real patient data.
- [x] 8.2 Verify full local startup from clean checkout using documented commands.
- [x] 8.3 Verify login, knowledge-base creation, document upload, ingestion completion, chat streaming, citations, and answer details end to end.
- [x] 8.4 Run frontend checks, backend tests, integration tests, and documented manual QA flows.
- [x] 8.5 Review every OpenSpec requirement and map it to tests or manual verification evidence.
- [x] 8.6 Update `.codex/wiki/` with stable architecture facts discovered during implementation.
- [x] 8.7 Archive or sync OpenSpec specs after implementation is accepted.
