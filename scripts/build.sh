#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

log "构建后端"
(cd "$BACKEND_DIR" && CGO_ENABLED=0 go build ./cmd/server)

log "安装前端依赖"
(cd "$FRONTEND_DIR" && npm install)

log "构建前端"
(cd "$FRONTEND_DIR" && npm run build)

log "构建完成"
