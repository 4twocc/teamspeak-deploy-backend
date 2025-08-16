#!/bin/bash

# TeamSpeak One-Click Deploy Backend Start Script
# 一键启动后端服务脚本

set -e  # 遇到错误时退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

print_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 显示使用说明
show_usage() {
    print_info "使用方法:"
    print_info "  ./start-backend.sh [选项]"
    print_info ""
    print_info "选项:"
    print_info "  -h, --help     显示此帮助信息"
    print_info "  -d, --daemon   后台运行服务"
    print_info "  -l, --log      指定日志文件 (默认: backend.log)"
    print_info ""
    print_info "示例:"
    print_info "  ./start-backend.sh              # 前台运行"
    print_info "  ./start-backend.sh -d           # 后台运行"
    print_info "  ./start-backend.sh -d -l app.log # 后台运行并指定日志文件"
}

# 默认参数
DAEMON=false
LOG_FILE="backend.log"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -d|--daemon)
            DAEMON=true
            shift
            ;;
        -l|--log)
            LOG_FILE="$2"
            shift 2
            ;;
        *)
            print_error "未知选项: $1"
            show_usage
            exit 1
            ;;
    esac
done

# 检查是否在正确的目录
if [ ! -f "backend/main.go" ]; then
    print_error "请在项目根目录运行此脚本"
    exit 1
fi

print_info "开始启动 TeamSpeak One-Click Deploy 后端服务..."

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    print_error "未找到 Go 环境，请先安装 Go 1.24+"
    exit 1
fi

# 检查 Go 版本
GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | cut -d ' ' -f 1)
print_info "当前 Go 版本: $GO_VERSION"

# 进入后端目录
cd backend

# 检查并安装依赖
print_info "检查和安装依赖..."
go mod tidy

# 检查环境变量文件
ENV_FILES=(".env" ".env.development" ".env.local")
ENV_FILE_FOUND=false

for env_file in "${ENV_FILES[@]}"; do
    if [ -f "$env_file" ]; then
        print_info "加载环境变量文件: $env_file"
        export $(grep -v '^#' $env_file | xargs)
        ENV_FILE_FOUND=true
        break
    fi
done

if [ "$ENV_FILE_FOUND" = false ]; then
    print_warn "未找到环境变量文件，将使用默认配置"
fi

# 创建必要的目录
print_info "创建必要的目录..."
mkdir -p data

# 构建应用
print_info "构建应用..."
go build -o teamspeak-backend main.go

cd ..

# 启动服务
if [ "$DAEMON" = true ]; then
    print_info "以后台模式启动服务..."
    print_info "日志文件: $LOG_FILE"
    nohup ./backend/teamspeak-backend > "$LOG_FILE" 2>&1 &
    PID=$!
    echo $PID > backend.pid
    print_info "服务已在后台启动，PID: $PID"
    print_info "使用 'kill $PID' 或 'kill \$(cat backend.pid)' 停止服务"
else
    print_info "启动后端服务..."
    print_info "服务将在端口 8080 上运行"
    print_info "按 Ctrl+C 可以停止服务"
    ./backend/teamspeak-backend
fi
