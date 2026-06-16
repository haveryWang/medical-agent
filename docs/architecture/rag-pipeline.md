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
  -> Generate embeddings with Volcengine Ark `doubao-embedding-vision-251215`
  -> Upsert vectors into Qdrant
  -> Persist chunk records and vector ids in MongoDB
  -> Mark document and knowledge base build status
```

## Chat Flow

```text
User sends question
  -> Persist user message
  -> Generate query embedding with Volcengine Ark `doubao-embedding-vision-251215`
  -> Search Qdrant with knowledge-base filters
  -> Hydrate chunks from MongoDB
  -> Assemble grounded prompt
  -> Call DeepSeek 对话 stream
  -> Stream SSE events to frontend
  -> Persist final assistant message, citations, and metadata
```

## Required Configuration

- `VOLCENGINE_API_KEY`: 火山引擎方舟 API Key，可同时作为 DeepSeek 对话和豆包向量模型的默认密钥。
- `DEEPSEEK_API_KEY`: backend-only DeepSeek credential for chat. If unset, falls back to `VOLCENGINE_API_KEY`.
- `DEEPSEEK_BASE_URL`: DeepSeek-compatible API base URL. Default: `https://ark.cn-beijing.volces.com/api/v3`.
- `DEEPSEEK_CHAT_MODEL`: chat model name. Default: `DeepSeek-V4-flash`.
- `QWEN_EMBEDDING_BASE_URL`: 火山引擎方舟向量模型 API base URL. Default: `https://ark.cn-beijing.volces.com/api/v3`.
- `QWEN_EMBEDDING_API_KEY`: backend-only embedding credential. If unset, falls back to `VOLCENGINE_API_KEY`.
- `QWEN_EMBEDDING_MODEL`: embedding model name. Default: `doubao-embedding-vision-251215`.
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
- If 火山引擎方舟向量模型 variables are not configured, the backend uses deterministic local vectors for development only.
- If DeepSeek is not configured, the backend streams a local demo answer for development only.
- MongoDB stores canonical metadata and message records. Qdrant stores vectors.
