# Subagent Handoffs

This document records planned subagent responsibilities for implementing the RAG knowledge-base chat system.

## Controller Agent

- Owns the OpenSpec change and implementation plan.
- Dispatches one implementation task at a time.
- Requires spec compliance review before code quality review.
- Ensures task records are updated.

## Product And Spec Agent

- Keeps OpenSpec requirements aligned with `design.png`.
- Reviews behavior against proposal, specs, and design.
- Flags scope drift.

## Frontend Agent

- Owns React project setup, routing, layout, login, chat workspace, knowledge-base management UI, upload panel, SSE rendering, and frontend runbook updates.

## Backend API Agent

- Owns Go server setup, auth/session APIs, REST handlers, middleware, error format, request ids, and backend runbook updates.

## Data And Storage Agent

- Owns MongoDB models, repositories, indexes, seed data, audit logs, Qdrant setup, and data model documentation.

## RAG Ingestion Agent

- Owns parsers, normalization, chunking, Qwen3-Embedding adapter, ingestion jobs, vector upsert, retries, and ingestion tests.

## RAG Chat Agent

- Owns Qwen3 query embedding, retrieval, prompt assembly, DeepSeek chat streaming adapter, SSE event generation, citations, answer detail data, and chat tests.

## QA And Review Agent

- Owns verification matrix, contract tests, end-to-end smoke tests, requirement-to-test mapping, and final review notes.

## Documentation Agent

- Owns API docs, architecture docs, startup/build docs, task records, and `.codex/wiki/` updates.
