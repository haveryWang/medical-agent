#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

log "运行后端测试"
(cd "$BACKEND_DIR" && CGO_ENABLED=0 go test ./...)

log "运行前端构建检查"
(cd "$FRONTEND_DIR" && npm install && npm run build)

log "验证完成"
