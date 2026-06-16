#!/usr/bin/env sh
# 跨平台开发环境检查（macOS / Linux / Git Bash on Windows）
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

MIN_GO_VERSION=1.22.0
MIN_NODE_MAJOR=18

log() {
  printf '\n==> %s\n' "$1"
}

warn() {
  printf '警告: %s\n' "$1" >&2
}

die() {
  printf '错误: %s\n' "$1" >&2
  exit 1
}

has_cmd() {
  command -v "$1" >/dev/null 2>&1
}

detect_os() {
  uname_s=$(uname -s 2>/dev/null || printf '')
  case "$uname_s" in
    Darwin*) OS=darwin ;;
    Linux*) OS=linux ;;
    MINGW*|MSYS*|CYGWIN*|Windows_NT*) OS=windows ;;
    *) OS=unknown ;;
  esac
  if [ -n "${MSYSTEM:-}" ] || [ -n "${WINDIR:-}" ]; then
    OS=windows
  fi
}

is_windows() {
  [ "${OS:-}" = "windows" ]
}

version_ge() {
  current=$1
  required=$2
  if has_cmd sort && printf '1.0\n1.1\n' | sort -CV >/dev/null 2>&1; then
    min=$(printf '%s\n%s' "$required" "$current" | sort -V | head -n 1)
    [ "$min" = "$required" ]
    return
  fi
  current_major=${current%%.*}
  current_minor=${current#*.}
  current_minor=${current_minor%%.*}
  required_major=${required%%.*}
  required_minor=${required#*.}
  required_minor=${required_minor%%.*}
  if [ "$current_major" -gt "$required_major" ]; then
    return 0
  fi
  [ "$current_major" -eq "$required_major" ] && [ "$current_minor" -ge "$required_minor" ]
}

go_version() {
  if ! has_cmd go; then
    return 1
  fi
  go env GOVERSION 2>/dev/null | sed 's/^go//'
}

node_major_version() {
  if ! has_cmd node; then
    return 1
  fi
  node -p "process.versions.node.split('.')[0]" 2>/dev/null
}

docker_ready() {
  has_cmd docker && docker info >/dev/null 2>&1
}

docker_compose_ready() {
  if has_cmd docker && docker compose version >/dev/null 2>&1; then
    return 0
  fi
  has_cmd docker-compose
}

install_hint() {
  tool=$1
  case "${OS:-unknown}" in
    darwin)
      printf '  macOS: brew install %s\n' "$2"
      ;;
    linux)
      printf '  Linux: 使用系统包管理器安装 %s，或参考官方文档\n' "$tool"
      ;;
    windows)
      printf '  Windows (PowerShell): powershell -ExecutionPolicy Bypass -File scripts/setup.ps1\n'
      printf '  Windows (winget): %s\n' "$3"
      ;;
    *)
      printf '  请参考 %s 官方安装文档\n' "$tool"
      ;;
  esac
}

try_install_go() {
  log "未检测到 Go >= ${MIN_GO_VERSION}，尝试自动安装"
  if is_windows && has_cmd winget; then
    winget install -e --id GoLang.Go --accept-package-agreements --accept-source-agreements
    return 0
  fi
  if [ "${OS:-}" = "darwin" ] && has_cmd brew; then
    brew install go
    return 0
  fi
  if [ "${OS:-}" = "linux" ] && has_cmd apt-get; then
    warn "将通过 apt 安装 golang-go，版本可能偏低，建议手动安装 Go ${MIN_GO_VERSION}+"
    sudo apt-get update && sudo apt-get install -y golang-go
    return 0
  fi
  return 1
}

try_install_node() {
  log "未检测到 Node.js >= ${MIN_NODE_MAJOR}，尝试自动安装"
  if is_windows && has_cmd winget; then
    winget install -e --id OpenJS.NodeJS.LTS --accept-package-agreements --accept-source-agreements
    return 0
  fi
  if [ "${OS:-}" = "darwin" ] && has_cmd brew; then
    brew install node
    return 0
  fi
  if [ "${OS:-}" = "linux" ] && has_cmd apt-get; then
    sudo apt-get update && sudo apt-get install -y nodejs npm
    return 0
  fi
  return 1
}

try_install_docker() {
  log "未检测到可用 Docker，尝试自动安装"
  if is_windows && has_cmd winget; then
    winget install -e --id Docker.DockerDesktop --accept-package-agreements --accept-source-agreements
    warn "Docker Desktop 安装后请启动应用，并等待引擎就绪后重试"
    return 0
  fi
  if [ "${OS:-}" = "darwin" ] && has_cmd brew; then
    brew install --cask docker
    warn "Docker Desktop 安装后请启动应用，并等待引擎就绪后重试"
    return 0
  fi
  if [ "${OS:-}" = "linux" ] && has_cmd apt-get; then
    warn "Linux 请按官方文档安装 Docker Engine: https://docs.docker.com/engine/install/"
    return 1
  fi
  return 1
}

ensure_go() {
  detect_os
  version=$(go_version || true)
  if [ -n "$version" ] && version_ge "$version" "$MIN_GO_VERSION"; then
    log "Go ${version} 已就绪"
    return 0
  fi

  if [ -n "$version" ]; then
    warn "当前 Go 版本 ${version} 低于要求的 ${MIN_GO_VERSION}"
  fi

  if [ "${ENSURE_AUTO_INSTALL:-1}" != "0" ]; then
    try_install_go || true
    version=$(go_version || true)
    if [ -n "$version" ] && version_ge "$version" "$MIN_GO_VERSION"; then
      log "Go ${version} 安装完成"
      return 0
    fi
  fi

  die "需要 Go >= ${MIN_GO_VERSION}。$(install_hint go 'go' 'winget install -e --id GoLang.Go')"
}

ensure_npm() {
  detect_os
  major=$(node_major_version || true)
  if [ -n "$major" ] && [ "$major" -ge "$MIN_NODE_MAJOR" ] && has_cmd npm; then
    npm_version=$(npm -v 2>/dev/null || printf '?')
    log "Node $(node -v 2>/dev/null || printf '?') / npm ${npm_version} 已就绪"
    return 0
  fi

  if [ -n "$major" ] && [ "$major" -lt "$MIN_NODE_MAJOR" ]; then
    warn "当前 Node 主版本 ${major} 低于要求的 ${MIN_NODE_MAJOR}"
  fi

  if [ "${ENSURE_AUTO_INSTALL:-1}" != "0" ]; then
    try_install_node || true
    major=$(node_major_version || true)
    if [ -n "$major" ] && [ "$major" -ge "$MIN_NODE_MAJOR" ] && has_cmd npm; then
      log "Node $(node -v) / npm $(npm -v) 安装完成"
      return 0
    fi
  fi

  die "需要 Node.js >= ${MIN_NODE_MAJOR} 与 npm。$(install_hint Node.js 'node' 'winget install -e --id OpenJS.NodeJS.LTS')"
}

ensure_docker() {
  detect_os
  if docker_ready && docker_compose_ready; then
    log "Docker 与 Docker Compose 已就绪"
    return 0
  fi

  if has_cmd docker && ! docker_ready; then
    warn "已安装 docker 命令，但 Docker 引擎未运行"
    if is_windows; then
      warn "请启动 Docker Desktop 后重试"
    elif [ "${OS:-}" = "darwin" ]; then
      warn "请启动 Docker Desktop 后重试"
    else
      warn "请执行: sudo systemctl start docker"
    fi
  fi

  if [ "${ENSURE_AUTO_INSTALL:-1}" != "0" ] && ! has_cmd docker; then
    try_install_docker || true
    if docker_ready && docker_compose_ready; then
      log "Docker 安装完成"
      return 0
    fi
  fi

  die "需要 Docker 与 Docker Compose。$(install_hint Docker 'docker' 'winget install -e --id Docker.DockerDesktop')"
}

ensure_dev_toolchain() {
  detect_os
  log "检查开发环境 (${OS})"
  ensure_go
  ensure_npm
  ensure_docker
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
