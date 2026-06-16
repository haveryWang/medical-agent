#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

log() {
  printf '\n==> %s\n' "$1"
}

ensure_backend_env() {
  if [ ! -f "$BACKEND_DIR/.env" ]; then
    cp "$BACKEND_DIR/.env.example" "$BACKEND_DIR/.env"
    log "已创建 backend/.env，请按需填写 VOLCENGINE_API_KEY"
  fi
}

ensure_frontend_env() {
  if [ ! -f "$FRONTEND_DIR/.env" ]; then
    cp "$FRONTEND_DIR/.env.example" "$FRONTEND_DIR/.env"
    log "已创建 frontend/.env"
  fi
}

load_backend_env() {
  ensure_backend_env
  set -a
  # shellcheck disable=SC1091
  . "$BACKEND_DIR/.env"
  set +a
}
