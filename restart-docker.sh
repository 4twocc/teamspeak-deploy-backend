#!/bin/bash

# TeamSpeak One-Click Deploy Restart Script with Docker
# 使用Docker重启服务脚本

set -e  # 遇到错误时退出

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

# 显示使用说明
show_usage() {
    print_info "使用方法:"
    print_info "  ./restart-docker.sh [选项]"
    print_info ""
    print_info "选项:"
    print_info "  -h, --help     显示此帮助信息"
    print_info "  -d, --dev      重启开发模式的服务"
    print_info "  -b, --build    重启时重新构建镜像"
    print_info ""
    print_info "示例:"
    print_info "  ./restart-docker.sh          # 重启生产模式服务"
    print_info "  ./restart-docker.sh -d       # 重启开发模式服务"
    print_info "  ./restart-docker.sh -b       # 重启并重新构建镜像"
}

# 默认参数
DEV_MODE=false
FORCE_BUILD=false

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
        -b|--build)
            FORCE_BUILD=true
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
    print_info "请在项目根目录运行此脚本"
    exit 1
fi

print_info "开始重启 TeamSpeak One-Click Deploy 服务..."

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
    print_info "重启开发模式服务"
else
    COMPOSE_FILE="docker-compose.yml"
    print_info "重启生产模式服务"
fi

# 检查 compose 文件是否存在
if [ ! -f "$COMPOSE_FILE" ]; then
    print_error "未找到 Docker Compose 文件: $COMPOSE_FILE"
    exit 1
fi

# 构建选项
BUILD_OPTION=""
if [ "$FORCE_BUILD" = true ]; then
    BUILD_OPTION="--build"
    print_info "将重新构建镜像"
fi

# 重启服务
print_info "重启服务..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE up $BUILD_OPTION -d --force-recreate

# 等待服务启动
print_info "等待服务启动..."
sleep 5

# 检查服务状态
print_info "检查服务状态..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE ps

print_info "服务重启完成！"
print_info "后端服务地址: http://localhost:8080"
print_info "Redis服务地址: redis://localhost:6379"
# if [ "$DEV_MODE" = true ]; then
#     print_info "前端开发服务器地址: http://localhost:3000"
# else
#     print_info "前端服务地址: http://localhost"
# fi
print_info "监控指标地址: http://localhost:9100"
