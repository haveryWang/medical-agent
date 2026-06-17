## Why

The project needs a complete medical knowledge-base conversation platform that can ingest internal documents, retrieve relevant knowledge with RAG, and answer user questions through a chat UI. The current repository only contains planning scaffolding, so this change defines the product, architecture, implementation tasks, and required documentation before code work begins.

## What Changes

- Add a React + JavaScript frontend with pages for login, agent conversation, knowledge-base list, document upload, and management navigation based on `design.png`.
- Add a Go backend that exposes authentication, conversation, knowledge-base, document ingestion, retrieval, and administration APIs.
- Add MongoDB persistence for users, sessions, conversations, messages, knowledge bases, documents, chunks, upload jobs, audit logs, and model configuration metadata.
- Add a lightweight vector database for document chunk embeddings and similarity search.
- Add RAG orchestration that retrieves relevant chunks, builds prompts, calls DeepSeek chat models, calls Qwen3-Embedding for vector generation, and streams responses to the frontend.
- Reserve separate configuration locations for the operator to supply DeepSeek chat credentials and Qwen3 embedding credentials/model settings without hardcoding secrets.
- Add complete documentation for startup, build, API integration, data model, RAG pipeline, streaming behavior, and implementation records.

## Capabilities

### New Capabilities

- `user-auth`: Login, session management, user identity, role-aware access, and logout.
- `agent-chat`: Conversation management, streaming chat responses, RAG answer generation, citations, and answer details.
- `knowledge-base-management`: Knowledge-base listing, filtering, metadata management, document upload, indexing lifecycle, and source management.
- `rag-ingestion-retrieval`: Document parsing, chunking, Qwen3-Embedding vector generation, vector indexing, retrieval strategy, prompt assembly, and DeepSeek chat integration.
- `system-documentation`: Required project documentation for setup, build, API contracts, architecture, operations, and task execution records.

### Modified Capabilities

- None. No accepted specs exist yet.

## Impact

- Adds frontend application source under a future `frontend/` project.
- Adds backend service source under a future `backend/` project.
- Adds local development dependencies for MongoDB, a lightweight vector database, and configuration files.
- Adds API contracts for frontend/backend integration, including streaming endpoints.
- Adds documentation under `docs/` and project knowledge notes under `.codex/wiki/`.
- Affects planning artifacts under `openspec/changes/build-rag-knowledge-chat-system/`.
