#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

load_backend_env

QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
QDRANT_COLLECTION="${QDRANT_COLLECTION:-medical_agent_chunks}"
QWEN_EMBEDDING_MODEL="${QWEN_EMBEDDING_MODEL:-doubao-embedding-vision-251215}"
QWEN_EMBEDDING_DIMENSION="${QWEN_EMBEDDING_DIMENSION:-2048}"
if [ "$QWEN_EMBEDDING_MODEL" = "doubao-embedding-vision-251215" ]; then
  QWEN_EMBEDDING_DIMENSION=2048
fi
MONGODB_DATABASE="${MONGODB_DATABASE:-medical_agent}"

mongo_eval() {
  script="$1"
  if command -v mongosh >/dev/null 2>&1; then
    mongosh "${MONGODB_URI}/${MONGODB_DATABASE}" --quiet --eval "$script"
    return
  fi
  docker exec medical-agent-mongodb mongosh "mongodb://localhost:27017/${MONGODB_DATABASE}" --quiet --eval "$script"
}

log "更新模型配置维度为 ${QWEN_EMBEDDING_DIMENSION}"
mongo_eval "
const now = new Date();
db.model_configs.updateMany(
  { qwenEmbeddingModel: 'doubao-embedding-vision-251215' },
  { \$set: { qwenEmbeddingDimension: Number(${QWEN_EMBEDDING_DIMENSION}), updatedAt: now } }
);
"

log "重建 Qdrant collection ${QDRANT_COLLECTION}"
curl -sS -X DELETE "${QDRANT_URL}/collections/${QDRANT_COLLECTION}" >/dev/null 2>&1 || true
(cd "$BACKEND_DIR" && sh scripts/qdrant-init.sh)

log "清理旧 chunks 并重新排队现有文档"
mongo_eval "
const now = new Date();
db.chunks.deleteMany({});
db.ingestion_jobs.deleteMany({});
const docs = db.documents.find({}, { _id: 1, knowledgeBaseId: 1 }).toArray();
if (docs.length > 0) {
  db.ingestion_jobs.insertMany(docs.map((doc) => ({
    knowledgeBaseId: doc.knowledgeBaseId,
    documentId: doc._id,
    status: 'pending',
    step: 'vector-index-reset',
    attempts: 0,
    createdAt: now,
    updatedAt: now
  })));
}
db.documents.updateMany({}, { \$set: { status: 'pending', failureReason: '', updatedAt: now } });
db.knowledge_bases.updateMany({}, { \$set: { chunkCount: 0, updatedAt: now } });
db.knowledge_bases.updateMany({ documentCount: { \$gt: 0 } }, { \$set: { buildStatus: 'building', updatedAt: now } });
print('重新排队文档数量: ' + docs.length);
"

log "向量索引已重置；启动后端后 worker 会重新向量化现有文档"
