#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

log "启动 MongoDB 和 Qdrant"
(cd "$ROOT_DIR" && docker compose up -d mongodb qdrant)

log "初始化 Qdrant collection"
load_backend_env
(cd "$BACKEND_DIR" && sh scripts/qdrant-init.sh)

log "容器状态"
(cd "$ROOT_DIR" && docker compose ps)
