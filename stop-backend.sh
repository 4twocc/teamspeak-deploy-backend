#!/bin/bash

# TeamSpeak One-Click Deploy Backend Stop Script
# 停止后台服务脚本

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的信息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否在正确的目录
if [ ! -f "backend/main.go" ]; then
    print_error "请在项目根目录运行此脚本"
    exit 1
fi

# 检查PID文件是否存在
if [ ! -f "backend.pid" ]; then
    print_warn "未找到 backend.pid 文件，服务可能未在运行或已使用其他方式启动"
    exit 1
fi

# 读取PID
PID=$(cat backend.pid)

# 检查进程是否存在
if ps -p $PID > /dev/null; then
    print_info "正在停止 PID 为 $PID 的服务..."
    kill $PID
    
    # 等待进程结束
    TIMEOUT=30
    COUNT=0
    while ps -p $PID > /dev/null && [ $COUNT -lt $TIMEOUT ]; do
        sleep 1
        COUNT=$((COUNT + 1))
    done
    
    if ps -p $PID > /dev/null; then
        print_warn "服务未能正常停止，强制终止..."
        kill -9 $PID
    fi
    
    # 删除PID文件
    rm -f backend.pid
    
    print_info "服务已停止"
else
    print_warn "PID 为 $PID 的进程不存在，可能服务已经停止"
    rm -f backend.pid
fi