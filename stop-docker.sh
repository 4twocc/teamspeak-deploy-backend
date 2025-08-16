#!/bin/bash

# TeamSpeak One-Click Deploy Stop Script with Docker
# 使用Docker停止服务脚本

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

# 显示使用说明
show_usage() {
    echo "使用方法:"
    echo "  ./stop-docker.sh [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -d, --dev      停止开发模式的服务"
    echo "  -v, --volumes  删除相关数据卷"
    echo ""
    echo "示例:"
    echo "  ./stop-docker.sh          # 停止生产模式服务"
    echo "  ./stop-docker.sh -d       # 停止开发模式服务"
    echo "  ./stop-docker.sh -v       # 停止服务并删除数据卷"
}

# 默认参数
DEV_MODE=false
REMOVE_VOLUMES=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -d|--dev)
            DEV_MODE=true
            shift
            ;;
        -v|--volumes)
            REMOVE_VOLUMES=true
            shift
            ;;
        *)
            print_error "未知选项: $1"
            show_usage
            exit 1
            ;;
    esac
done

# 检查是否在正确的目录
if [ ! -f "docker-compose.yml" ]; then
    print_error "请在项目根目录运行此脚本"
    exit 1
fi

print_info "开始停止 TeamSpeak One-Click Deploy 服务..."

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    print_error "未找到 Docker，请先安装 Docker"
    exit 1
fi

# 检查 Docker Compose 是否安装
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    print_error "未找到 Docker Compose，请先安装 Docker Compose"
    exit 1
fi

# 根据模式选择 docker-compose 文件
if [ "$DEV_MODE" = true ]; then
    COMPOSE_FILE="docker-compose.dev.yml"
    print_info "停止开发模式服务"
else
    COMPOSE_FILE="docker-compose.yml"
    print_info "停止生产模式服务"
fi

# 检查 compose 文件是否存在
if [ ! -f "$COMPOSE_FILE" ]; then
    print_error "未找到 Docker Compose 文件: $COMPOSE_FILE"
    exit 1
fi

# 停止选项
DOWN_OPTION=""
if [ "$REMOVE_VOLUMES" = true ]; then
    DOWN_OPTION="--volumes"
    print_warn "将删除相关数据卷（数据将丢失）"
fi

# 停止服务
print_info "停止服务..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down $DOWN_OPTION

print_info "服务已停止"