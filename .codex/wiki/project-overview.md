# Project Overview

## Identity

- Repository name: `medical-agent`
- Current visible source: project documentation and OpenSpec scaffolding.
- Current OpenSpec schema: `spec-driven`

## Current State

The repository contains an implemented full-stack RAG knowledge-base
conversation system:

- Go backend under `backend/`
- React frontend under `frontend/`
- MongoDB for canonical data
- Qdrant for vector search
- Volcengine Ark-compatible model settings stored in MongoDB

The active OpenSpec change for this implementation remains:

- `openspec/changes/build-rag-knowledge-chat-system/`

## Canonical Sources

- `README.md` contains the public project name and local startup guidance.
- `design.png` contains the initial UI and system requirement design source.
- `openspec/config.yaml` defines the OpenSpec workflow configuration.
- `openspec/specs/` is reserved for accepted capability specs.
- `openspec/changes/` is reserved for proposed or in-progress changes.
- `docs/superpowers/plans/2026-06-15-rag-knowledge-chat-system.md` contains
  the implementation plan for the active RAG system change.
- `docs/records/task-log.md` is the required implementation record.

## Known Runtime Notes

- Backend tests run with `CGO_ENABLED=0 go test ./...` on this macOS setup to
  avoid the local dyld `missing LC_UUID load command` failure.
- Frontend production build runs from `frontend/` with `npm run build`.
