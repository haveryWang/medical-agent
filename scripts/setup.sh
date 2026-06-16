#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

log "准备环境变量文件"
ensure_backend_env
ensure_frontend_env

log "安装前端依赖"
(cd "$FRONTEND_DIR" && npm install)

log "下载后端依赖"
(cd "$BACKEND_DIR" && go mod download)

log "准备完成"
printf '后端配置：%s\n' "$BACKEND_DIR/.env"
printf '前端配置：%s\n' "$FRONTEND_DIR/.env"
