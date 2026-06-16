#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

ensure_frontend_env
ensure_npm

if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
  log "安装前端依赖"
  (cd "$FRONTEND_DIR" && npm install)
fi

log "启动前端 http://localhost:5173"
(cd "$FRONTEND_DIR" && npm run dev)
