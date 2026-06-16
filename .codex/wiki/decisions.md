# Decisions

Record durable technical and process decisions here. Link to related OpenSpec
changes when possible.

## 2026-06-15: Initialize Codex Wiki

- Decision: use `.codex/wiki/` as the repository-local Codex knowledge base.
- Decision: add root `AGENTS.md` so future Codex sessions know to read the wiki.
- Reason: the repository had OpenSpec scaffolding but no persistent project
  notes for agent sessions.

## 2026-06-15: Plan RAG Knowledge-Base Chat System

- Decision: create OpenSpec change `build-rag-knowledge-chat-system`.
- Decision: use React + JavaScript for frontend and Go for backend.
- Decision: use MongoDB as the source of truth for non-vector data.
- Decision: use Qdrant as the default lightweight vector database behind an
  adapter.
- Decision: use DeepSeek for chat generation through backend-only
  configuration, with `DEEPSEEK_API_KEY` left as an operator-supplied secret.
- Decision: use Qwen3-Embedding for document and query vector generation, with
  separate `QWEN_EMBEDDING_*` configuration.
- Decision: use SSE for initial streaming chat responses.
- Reason: DeepSeek handles conversation quality while Qwen3-Embedding is a
  better fit for embedding; keeping separate configuration makes model
  operations explicit.
