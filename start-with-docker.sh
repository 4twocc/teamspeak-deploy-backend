#!/bin/bash

# TeamSpeak One-Click Deploy Start Script with Docker
# 使用Docker一键启动服务脚本

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
    echo "使用方法:"
    echo "  ./start-with-docker.sh [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -d, --dev      使用开发模式启动 (支持热重载)"
    echo "  -b, --build    强制重新构建镜像"
    echo ""
    echo "示例:"
    echo "  ./start-with-docker.sh          # 生产模式启动"
    echo "  ./start-with-docker.sh -d       # 开发模式启动"
    echo "  ./start-with-docker.sh -b       # 强制重新构建并启动"
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
    print_error "请在项目根目录运行此脚本"
    exit 1
fi

print_info "开始使用 Docker 启动 TeamSpeak One-Click Deploy 服务..."

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    print_error "未找到 Docker，请先安装 Docker"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null; then
    print_error "未找到 Docker Compose，请先安装 Docker Compose"
    exit 1
fi

print_info "Docker 环境检查通过"

# 检查环境变量文件
ENV_FILES=("backend/.env" "backend/.env.development" "backend/.env.production")
ENV_FILE_FOUND=false

for env_file in "${ENV_FILES[@]}"; do
    if [ -f "$env_file" ]; then
        print_info "找到环境变量文件: $env_file"
        ENV_FILE_FOUND=true
        break
    fi
done

if [ "$ENV_FILE_FOUND" = false ]; then
    print_warn "未找到环境变量文件，将使用默认配置"
fi

# 创建必要的目录
print_info "创建必要的目录..."
mkdir -p backend/data

# 设置正确的权限
chmod 755 backend/data

# 根据模式选择 docker-compose 文件
if [ "$DEV_MODE" = true ]; then
    COMPOSE_FILE="docker-compose.dev.yml"
    print_info "使用开发模式启动服务"
else
    COMPOSE_FILE="docker-compose.yml"
    print_info "使用生产模式启动服务"
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
    print_info "强制重新构建镜像"
fi

# 启动服务
print_info "启动服务..."
if [ "$DEV_MODE" = true ]; then
    docker-compose -f $COMPOSE_FILE up $BUILD_OPTION -d
else
    docker-compose -f $COMPOSE_FILE up $BUILD_OPTION -d
fi

# 等待服务启动
print_info "等待服务启动..."
sleep 5

# 检查服务状态
print_info "检查服务状态..."
docker-compose -f $COMPOSE_FILE ps

print_info "服务启动完成！"
print_info "后端服务地址: http://localhost:8080"
if [ "$DEV_MODE" = true ]; then
    print_info "前端开发服务器地址: http://localhost:3000"
else
    print_info "前端服务地址: http://localhost"
fi
print_info "监控指标地址: http://localhost:9100"

print_info "使用以下命令查看日志:"
print_info "  docker-compose -f $COMPOSE_FILE logs -f"
print_info ""
print_info "使用以下命令停止服务:"
print_info "  docker-compose -f $COMPOSE_FILE down"