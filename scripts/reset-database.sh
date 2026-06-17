#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

if [ "${1:-}" != "--yes" ]; then
  printf '此操作会清空 MongoDB 业务数据和 Qdrant 向量集合，仅保留/重建 admin 账号。\n'
  printf '确认请输入 RESET：'
  read -r CONFIRM
  if [ "$CONFIRM" != "RESET" ]; then
    printf '已取消。\n'
    exit 1
  fi
fi

load_backend_env

log "确保数据库容器已启动"
(cd "$ROOT_DIR" && docker compose up -d mongodb qdrant)

log "清空 MongoDB 业务数据，仅保留 admin"
(cd "$ROOT_DIR" && docker compose exec -T mongodb mongosh "$MONGODB_DATABASE" --quiet <<'MONGOSH'
const now = new Date();
const adminHash = "$2a$10$9RHOKRbanMlsG5HNl/ebUuGfuYbtTjnTgb.pSxVY.LuXMj2wPMb1K";
const admin = db.users.findOne({ account: "admin" });

db.sessions.deleteMany({});
db.roles.deleteMany({});
db.knowledge_bases.deleteMany({});
db.documents.deleteMany({});
db.chunks.deleteMany({});
db.ingestion_jobs.deleteMany({});
db.conversations.deleteMany({});
db.messages.deleteMany({});
db.audit_logs.deleteMany({});
db.review_notes.deleteMany({});
db.review_note_exports.deleteMany({});
db.policy_documents.deleteMany({});
db.policy_import_batches.deleteMany({});
db.model_configs.deleteMany({});
db.users.deleteMany({ account: { $ne: "admin" } });

if (admin) {
  db.users.updateOne(
    { _id: admin._id },
    {
      $set: {
        account: "admin",
        passwordHash: adminHash,
        displayName: "系统管理员",
        roles: ["系统管理员", "知识库管理员"],
        permissions: ["chat:use", "knowledge:read", "knowledge:write", "review_notes:write", "policy:write", "system:read"],
        status: "active",
        updatedAt: now
      },
      $setOnInsert: { createdAt: now }
    }
  );
} else {
  db.users.insertOne({
    _id: ObjectId(),
    account: "admin",
    passwordHash: adminHash,
    displayName: "系统管理员",
    roles: ["系统管理员", "知识库管理员"],
    permissions: ["chat:use", "knowledge:read", "knowledge:write", "review_notes:write", "policy:write", "system:read"],
    status: "active",
    createdAt: now,
    updatedAt: now
  });
}

printjson({
  users: db.users.countDocuments({}),
  admin: db.users.countDocuments({ account: "admin", status: "active" }),
  knowledge_bases: db.knowledge_bases.countDocuments({}),
  documents: db.documents.countDocuments({}),
  chunks: db.chunks.countDocuments({}),
  review_notes: db.review_notes.countDocuments({}),
  review_note_exports: db.review_note_exports.countDocuments({}),
  policy_documents: db.policy_documents.countDocuments({}),
  policy_import_batches: db.policy_import_batches.countDocuments({}),
  conversations: db.conversations.countDocuments({}),
  messages: db.messages.countDocuments({})
});
MONGOSH
)

log "重建 Qdrant collection"
curl -sS -X DELETE "${QDRANT_URL}/collections/${QDRANT_COLLECTION}" >/dev/null 2>&1 || true
(cd "$BACKEND_DIR" && sh scripts/qdrant-init.sh)

log "数据库已重置"
printf '保留账号：admin / admin123\n'
