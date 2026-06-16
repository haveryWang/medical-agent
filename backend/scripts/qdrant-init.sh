#!/usr/bin/env sh
set -eu

QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
QDRANT_COLLECTION="${QDRANT_COLLECTION:-medical_agent_chunks}"
QWEN_EMBEDDING_MODEL="${QWEN_EMBEDDING_MODEL:-doubao-embedding-vision-251215}"
QWEN_EMBEDDING_DIMENSION="${QWEN_EMBEDDING_DIMENSION:-2048}"
if [ "$QWEN_EMBEDDING_MODEL" = "doubao-embedding-vision-251215" ]; then
  QWEN_EMBEDDING_DIMENSION=2048
fi

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
