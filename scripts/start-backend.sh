#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
# shellcheck disable=SC1091
. "$SCRIPT_DIR/lib.sh"

load_backend_env

log "启动后端 http://localhost:8080"
(cd "$BACKEND_DIR" && CGO_ENABLED=0 go run ./cmd/server)
