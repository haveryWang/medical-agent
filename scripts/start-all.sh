#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

cleanup() {
  if [ "${BACKEND_PID:-}" ]; then
    kill "$BACKEND_PID" 2>/dev/null || true
  fi
  if [ "${FRONTEND_PID:-}" ]; then
    kill "$FRONTEND_PID" 2>/dev/null || true
  fi
}
trap cleanup INT TERM EXIT

sh "$SCRIPT_DIR/start-db.sh"

log "后台启动后端"
sh "$SCRIPT_DIR/start-backend.sh" &
BACKEND_PID=$!

log "后台启动前端"
sh "$SCRIPT_DIR/start-frontend.sh" &
FRONTEND_PID=$!

log "服务已启动"
printf '前端：http://localhost:5173\n'
printf '后端：http://localhost:8080\n'
printf '按 Ctrl+C 停止前端和后端进程；数据库容器会继续运行。\n'

wait
