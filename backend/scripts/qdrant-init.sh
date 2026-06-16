#!/usr/bin/env sh
set -eu

QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
QDRANT_COLLECTION="${QDRANT_COLLECTION:-medical_agent_chunks}"
QWEN_EMBEDDING_DIMENSION="${QWEN_EMBEDDING_DIMENSION:-1024}"

curl -sS -X PUT "${QDRANT_URL}/collections/${QDRANT_COLLECTION}" \
  -H "Content-Type: application/json" \
  -d "{
    \"vectors\": {
      \"size\": ${QWEN_EMBEDDING_DIMENSION},
      \"distance\": \"Cosine\"
    }
  }"

echo
echo "Qdrant collection ready: ${QDRANT_COLLECTION}"
