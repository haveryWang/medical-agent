# RAG Pipeline

## Ingestion Flow

```text
Upload document
  -> Validate file type and size
  -> Store document metadata in MongoDB
  -> Create ingestion job
  -> Parse document text
  -> Normalize text
  -> Chunk text with overlap
  -> Generate embeddings with Qwen3-Embedding
  -> Upsert vectors into Qdrant
  -> Persist chunk records and vector ids in MongoDB
  -> Mark document and knowledge base build status
```

## Chat Flow

```text
User sends question
  -> Persist user message
  -> Generate query embedding with Qwen3-Embedding
  -> Search Qdrant with knowledge-base filters
  -> Hydrate chunks from MongoDB
  -> Assemble grounded prompt
  -> Call DeepSeek chat stream
  -> Stream SSE events to frontend
  -> Persist final assistant message, citations, and metadata
```

## Required Configuration

- `DEEPSEEK_API_KEY`: backend-only DeepSeek credential for chat.
- `DEEPSEEK_BASE_URL`: DeepSeek-compatible API base URL.
- `DEEPSEEK_CHAT_MODEL`: chat model name. Default planned value: `deepseek-v4-flash`; use `deepseek-v4-pro` if higher reasoning quality is required.
- `QWEN_EMBEDDING_BASE_URL`: Qwen3-Embedding-compatible API base URL.
- `QWEN_EMBEDDING_API_KEY`: backend-only Qwen embedding credential.
- `QWEN_EMBEDDING_MODEL`: embedding model name. Planned value: `Qwen3-Embedding`.
- `QWEN_EMBEDDING_DIMENSION`: vector dimension used when creating the Qdrant collection.
- `MONGODB_URI`: MongoDB connection string.
- `QDRANT_URL`: Qdrant connection URL.

## Grounding Rules

- Answers must prefer retrieved knowledge-base context over model priors.
- Answers must expose citations when chunks are used.
- If no reliable context is found, the system must say so instead of inventing sources.
- Answer details must preserve retrieval metadata for audit and troubleshooting.

## Current Implementation Notes

- Uploaded files are stored under `UPLOAD_DIR`, default `../data/uploads`.
- The demo parser reads text-like content directly and keeps metadata for binary files. Production parsing adapters can replace `backend/internal/ingestion/ingestion.go`.
- If Qwen3-Embedding variables are not configured, the backend uses deterministic local vectors for development only.
- If DeepSeek is not configured, the backend streams a local demo answer for development only.
- MongoDB stores canonical metadata and message records. Qdrant stores vectors.
