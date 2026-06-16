# RAG Knowledge-Base Chat System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete React + Go RAG knowledge-base conversation system with login, streaming agent chat, knowledge-base management, MongoDB persistence, lightweight vector search, DeepSeek chat integration, Qwen3-Embedding vector generation, and full documentation records.

**Architecture:** The frontend is a JavaScript React app organized by auth, chat, knowledge, and shared modules. The backend is a Go HTTP service with layered handlers, services, MongoDB repositories, Qdrant vector adapter, DeepSeek chat adapter, Qwen3-Embedding adapter, and asynchronous ingestion worker. Documentation is a first-class deliverable and must be updated in every task.

**Tech Stack:** JavaScript, React, Go, MongoDB, Qdrant, DeepSeek chat API, Qwen3-Embedding, SSE streaming.

---

## Source Documents

- OpenSpec proposal: `openspec/changes/build-rag-knowledge-chat-system/proposal.md`
- OpenSpec design: `openspec/changes/build-rag-knowledge-chat-system/design.md`
- OpenSpec tasks: `openspec/changes/build-rag-knowledge-chat-system/tasks.md`
- Specs:
  - `openspec/changes/build-rag-knowledge-chat-system/specs/user-auth/spec.md`
  - `openspec/changes/build-rag-knowledge-chat-system/specs/agent-chat/spec.md`
  - `openspec/changes/build-rag-knowledge-chat-system/specs/knowledge-base-management/spec.md`
  - `openspec/changes/build-rag-knowledge-chat-system/specs/rag-ingestion-retrieval/spec.md`
  - `openspec/changes/build-rag-knowledge-chat-system/specs/system-documentation/spec.md`
- UI design source: `design.png`
- Handoff records: `docs/records/subagent-handoffs.md`
- Task log: `docs/records/task-log.md`

## Subagent Work Allocation

| Workstream | Primary Role | Key Files |
| --- | --- | --- |
| Product/spec | Product And Spec Agent | `openspec/changes/build-rag-knowledge-chat-system/**` |
| Frontend | Frontend Agent | `frontend/**`, `docs/runbooks/frontend.md` |
| Backend API | Backend API Agent | `backend/**`, `docs/runbooks/backend.md`, `docs/api/**` |
| Data/storage | Data And Storage Agent | `backend/internal/storage/**`, MongoDB indexes, Qdrant setup docs |
| RAG ingestion | RAG Ingestion Agent | `backend/internal/ingestion/**`, `backend/internal/rag/**`, `backend/internal/providers/qwen/**`, `docs/architecture/rag-pipeline.md` |
| RAG chat | RAG Chat Agent | `backend/internal/chat/**`, `backend/internal/providers/deepseek/**`, `frontend/chat/**` |
| QA/review | QA And Review Agent | tests, verification matrix, requirement mapping |
| Documentation | Documentation Agent | `docs/**`, `.codex/wiki/**` |

Every task must append a task record to `docs/records/task-log.md`.

## Task 1: Documentation Baseline Review

**Files:**
- Read: `design.png`
- Read: `openspec/changes/build-rag-knowledge-chat-system/proposal.md`
- Read: `openspec/changes/build-rag-knowledge-chat-system/design.md`
- Read: `openspec/changes/build-rag-knowledge-chat-system/specs/**/*.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Verify scope alignment**

Check that login, agent chat, knowledge-base list, upload panel, streaming output, citations, answer detail panel, MongoDB, Qdrant, DeepSeek chat, Qwen3-Embedding, and complete documentation are represented in OpenSpec.

- [ ] **Step 2: Record review result**

Append a task-log entry:

```markdown
## 2026-06-15 - Task 1 - Documentation Baseline Review

- Owner role: Product And Spec Agent
- Related OpenSpec requirement: system-documentation
- Files changed: docs/records/task-log.md
- Commands run: openspec validate build-rag-knowledge-chat-system --strict
- Verification result: Pending until validation is run during implementation
- Documentation updated: Task log
- Follow-up items: Fill with any scope gaps found
```

- [ ] **Step 3: Validate OpenSpec**

Run:

```bash
openspec validate build-rag-knowledge-chat-system --strict
```

Expected: validation passes. If it fails, fix specs before application scaffolding.

## Task 2: Frontend Project Scaffold

**Files:**
- Create: `frontend/`
- Modify: `docs/runbooks/frontend.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Scaffold React app**

Create a JavaScript React app under `frontend/`. Use a mainstream toolchain that supports local dev server and production build.

- [ ] **Step 2: Add route skeleton**

Create route placeholders for:

- `/login`
- `/chat`
- `/knowledge-bases`
- `/knowledge-bases/upload`

- [ ] **Step 3: Add frontend environment template**

Create a frontend env template with:

```text
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

- [ ] **Step 4: Update frontend runbook**

Document exact install, dev, build, lint, and test commands in `docs/runbooks/frontend.md`.

- [ ] **Step 5: Verify frontend**

Run the documented frontend checks and record results in `docs/records/task-log.md`.

## Task 3: Backend Project Scaffold

**Files:**
- Create: `backend/`
- Modify: `docs/runbooks/backend.md`
- Modify: `docs/api/rest-api.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Initialize Go module**

Create `backend/` with a Go module and server entrypoint at `backend/cmd/server`.

- [ ] **Step 2: Add configuration keys**

Backend configuration must include:

```text
HTTP_ADDR
MONGODB_URI
MONGODB_DATABASE
QDRANT_URL
DEEPSEEK_BASE_URL
DEEPSEEK_CHAT_MODEL
DEEPSEEK_API_KEY
QWEN_EMBEDDING_BASE_URL
QWEN_EMBEDDING_API_KEY
QWEN_EMBEDDING_MODEL
QWEN_EMBEDDING_DIMENSION
```

- [ ] **Step 3: Add health endpoint**

Expose a backend health endpoint that returns service status.

- [ ] **Step 4: Update backend runbook**

Document exact Go commands and local dependency setup.

- [ ] **Step 5: Verify backend**

Run Go tests/build and record results in `docs/records/task-log.md`.

## Task 4: Local Infrastructure

**Files:**
- Create or modify: local development dependency configuration
- Modify: `docs/runbooks/backend.md`
- Modify: `docs/architecture/overview.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Add MongoDB local setup**

Document and provide repeatable local MongoDB startup.

- [ ] **Step 2: Add Qdrant local setup**

Document and provide repeatable local Qdrant startup.

- [ ] **Step 3: Add backend environment example**

Ensure DeepSeek API key is represented as a placeholder only:

```text
DEEPSEEK_API_KEY=replace-with-your-key
```

- [ ] **Step 4: Verify dependencies**

Run local health checks for MongoDB and Qdrant and record results.

## Task 5: Authentication

**Files:**
- Create/modify: `backend/internal/auth/**`
- Create/modify: `backend/internal/storage/mongo/**`
- Create/modify: `frontend/src/auth/**`
- Modify: `docs/api/rest-api.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Backend auth models and repositories**

Implement users, roles, permissions, sessions, password hashing, and session validation.

- [ ] **Step 2: Backend auth APIs**

Implement:

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

- [ ] **Step 3: Frontend login**

Implement the login page matching `design.png`, protected route behavior, and logout.

- [ ] **Step 4: Tests**

Verify successful login, invalid login, missing session, expired session, and role-aware navigation.

- [ ] **Step 5: Documentation**

Update API docs and task log with commands and verification evidence.

## Task 6: Knowledge-Base Management

**Files:**
- Create/modify: `backend/internal/knowledge/**`
- Create/modify: `backend/internal/ingestion/**`
- Create/modify: `frontend/src/knowledge/**`
- Modify: `docs/api/rest-api.md`
- Modify: `docs/architecture/overview.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Backend metadata APIs**

Implement knowledge-base list, filters, create, edit, disable, document list, upload, ingestion status, and retry APIs.

- [ ] **Step 2: Frontend management UI**

Implement sidebar navigation, filters, list table, upload panel, validation, and status display matching `design.png`.

- [ ] **Step 3: Audit events**

Record audit events for management operations.

- [ ] **Step 4: Tests**

Verify filters, pagination, file validation, upload success/rejection, ingestion status, retry, and audit records.

- [ ] **Step 5: Documentation**

Update API docs, architecture docs, and task log.

## Task 7: RAG Ingestion And Vector Indexing

**Files:**
- Create/modify: `backend/internal/ingestion/**`
- Create/modify: `backend/internal/rag/**`
- Create/modify: `backend/internal/vector/**`
- Create/modify: `backend/internal/providers/qwen/**`
- Modify: `docs/architecture/rag-pipeline.md`
- Modify: `docs/runbooks/backend.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Parsers**

Implement parser interfaces and initial parsing support for PDF, Word, Excel, Markdown, and text.

- [ ] **Step 2: Chunking**

Implement normalization, chunk size, overlap, metadata, and checksums.

- [ ] **Step 3: Embeddings**

Implement Qwen3-Embedding adapter, separate configuration loading, and vector dimension validation.

- [ ] **Step 4: Vector store**

Implement Qdrant collection setup, vector upsert, filtered search, and delete/exclude behavior.

- [ ] **Step 5: Worker**

Implement asynchronous ingestion job processing and retry.

- [ ] **Step 6: Tests and docs**

Verify parser failures, chunking, missing Qwen3 embedding configuration, unavailable Qwen3-Embedding model, vector dimension mismatch, vector upsert, retry, and Mongo/vector consistency. Update docs.

## Task 8: Agent Chat And Streaming

**Files:**
- Create/modify: `backend/internal/chat/**`
- Create/modify: `backend/internal/rag/**`
- Create/modify: `backend/internal/providers/deepseek/**`
- Create/modify: `frontend/src/chat/**`
- Modify: `docs/api/streaming-chat.md`
- Modify: `docs/architecture/rag-pipeline.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Conversation APIs**

Implement conversation list, creation, message history, and answer detail APIs.

- [ ] **Step 2: Retrieval and prompt assembly**

Implement Qwen3 query embedding, Qdrant search, MongoDB chunk hydration, context ranking, prompt assembly, and no-context fallback.

- [ ] **Step 3: DeepSeek provider**

Implement backend-only DeepSeek chat streaming adapter using `DEEPSEEK_API_KEY`.

- [ ] **Step 4: SSE endpoint**

Implement streaming endpoint events:

- `message.started`
- `retrieval.sources`
- `message.delta`
- `message.completed`
- `message.error`

- [ ] **Step 5: Frontend chat UI**

Implement conversation list, search, message composer, progressive answer rendering, citations, message actions, and answer detail drawer matching `design.png`.

- [ ] **Step 6: Tests and docs**

Verify streaming success, failure, missing key, no-context fallback, citation persistence, and detail rendering. Update docs.

## Task 9: Security, Observability, And Operations

**Files:**
- Create/modify: backend middleware and operational packages
- Modify: `docs/runbooks/backend.md`
- Modify: `docs/architecture/overview.md`
- Modify: `docs/records/task-log.md`

- [ ] **Step 1: Request observability**

Add request ids, structured logs, safe errors, and timing metadata.

- [ ] **Step 2: Security controls**

Add upload limits, type validation, authorization checks, session expiry, secret redaction, and forbidden states.

- [ ] **Step 3: Indexes and health checks**

Add MongoDB indexes and health checks for backend, MongoDB, Qdrant, DeepSeek chat configuration, and Qwen3-Embedding configuration.

- [ ] **Step 4: Documentation**

Document health checks, common failures, and recovery steps.

## Task 10: Full-System Verification

**Files:**
- Modify: `docs/records/task-log.md`
- Modify: `docs/records/subagent-handoffs.md`
- Modify: `.codex/wiki/*.md`
- Modify: OpenSpec artifacts if requirement gaps are found

- [ ] **Step 1: Demo data**

Create safe demo data without real patient information.

- [ ] **Step 2: Clean startup verification**

From a clean checkout, follow documented commands to start MongoDB, Qdrant, backend, and frontend.

- [ ] **Step 3: End-to-end QA**

Verify login, knowledge-base creation, upload, ingestion, chat streaming, citations, and answer detail panel.

- [ ] **Step 4: Requirement mapping**

Map every OpenSpec requirement to automated tests or manual verification evidence.

- [ ] **Step 5: Handoff**

Update wiki with stable architecture facts and prepare OpenSpec sync/archive after acceptance.

## Execution Gate

This plan is ready for implementation after the user exits explore mode and asks to implement/apply the change. Implementation should use `openspec-apply-change` plus `subagent-driven-development`, one task at a time, with spec review and code quality review after each task.
