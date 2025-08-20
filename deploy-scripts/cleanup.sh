#!/bin/bash

# 清理脚本
# 用于停止和删除TeamSpeak服务及相关数据

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 打印带颜色的信息
print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_debug() {
    echo -e "${PURPLE}[DEBUG]${NC} $1"
}

set -e

print_warn "Cleaning up TeamSpeak deployment..."

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]
  then print_warn "Please run as root"
  exit 1
fi

# 停止并删除容器
if docker ps -a --format '{{.Names}}' | grep -q '^teamspeak-main$'; then
    print_warn "Stopping TeamSpeak container..."
    docker stop teamspeak-main || true
    print_warn "Removing TeamSpeak container..."
    docker rm teamspeak-main || true
else
    print_info "No TeamSpeak container found"
fi

# 停止并删除所有通过docker compose启动的服务
if [ -f "docker-compose.yml" ]; then
    print_info "Stopping services via docker compose..."
    docker compose down || true
fi

# 删除数据目录（可选，用户可以选择保留数据）
print_warn "Note: Data directory /var/lib/teamspeak is preserved."
print_success "To remove it completely, run: rm -rf /var/lib/teamspeak"

print_success "Cleanup completed successfully"