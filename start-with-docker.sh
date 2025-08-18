#!/bin/bash

# TeamSpeak One-Click Deploy Start Script with Docker
# 使用Docker一键启动服务脚本

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
    print_info "  ./start-with-docker.sh [选项]"
    print_info ""
    print_info "选项:"
    print_info "  -h, --help       显示此帮助信息"
    print_info "  -d, --dev        使用开发模式启动 (支持热重载)"
    print_info "  -b, --build      强制重新构建镜像"
    print_info "  -r, --registry   配置 Docker 镜像源"
    print_info "  -p, --pnpm       安装并配置 pnpm"
    print_info ""
    print_info "示例:"
    print_info "  ./start-with-docker.sh             # 生产模式启动"
    print_info "  ./start-with-docker.sh -d          # 开发模式启动"
    print_info "  ./start-with-docker.sh -b          # 强制重新构建并启动"
    print_info "  ./start-with-docker.sh -r          # 配置 Docker 镜像源并启动"
    print_info "  ./start-with-docker.sh -p          # 安装并配置 pnpm 并启动"
    print_info "  ./start-with-docker.sh -r -p       # 配置 Docker 镜像源并安装配置 pnpm"
    print_info "  ./start-with-docker.sh -r -d       # 配置 Docker 镜像源并以开发模式启动"
    print_info "  ./start-with-docker.sh -p -d       # 安装配置 pnpm 并以开发模式启动"
}

# 配置 Docker 镜像源
configure_docker_registry() {
    print_info "配置 Docker 镜像加速器..."
    
    # 检查是否以 root 权限运行
    if [ "$EUID" -ne 0 ] && ! command -v sudo &> /dev/null; then
        print_warn "未找到 sudo 命令且当前不是 root 用户，跳过 Docker 镜像源配置"
        return
    fi
    
    # 创建目录
    if command -v sudo &> /dev/null; then
        sudo mkdir -p /etc/docker
    else
        mkdir -p /etc/docker
    fi
    
    # 配置镜像源（强制更新配置）
    if command -v sudo &> /dev/null; then
        sudo tee /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://docker.registry.cyou",
    "https://docker-cf.registry.cyou",
    "https://dockercf.jsdelivr.fyi",
    "https://docker.jsdelivr.fyi",
    "https://dockertest.jsdelivr.fyi",
    "https://mirror.aliyuncs.com",
    "https://dockerproxy.com",
    "https://mirror.baidubce.com",
    "https://docker.m.daocloud.io",
    "https://docker.nju.edu.cn",
    "https://docker.mirrors.sjtug.sjtu.edu.cn",
    "https://docker.mirrors.ustc.edu.cn",
    "https://mirror.iscas.ac.cn",
    "https://docker.rainbond.cc",
    "https://do.nark.eu.org",
    "https://dc.j8.work",
    "https://dockerproxy.com",
    "https://gst6rzl9.mirror.aliyuncs.com",
    "https://registry.docker-cn.com",
    "http://hub-mirror.c.163.com",
    "http://mirrors.ustc.edu.cn/",
    "https://mirrors.tuna.tsinghua.edu.cn/",
    "http://mirrors.sohu.com/"
  ],
  "insecure-registries": [
    "registry.docker-cn.com",
    "docker.mirrors.ustc.edu.cn"
  ],
  "debug": true,
  "experimental": false
}
EOF
        sudo systemctl daemon-reload
        sudo systemctl restart docker
    else
        tee /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://docker.registry.cyou",
    "https://docker-cf.registry.cyou",
    "https://dockercf.jsdelivr.fyi",
    "https://docker.jsdelivr.fyi",
    "https://dockertest.jsdelivr.fyi",
    "https://mirror.aliyuncs.com",
    "https://dockerproxy.com",
    "https://mirror.baidubce.com",
    "https://docker.m.daocloud.io",
    "https://docker.nju.edu.cn",
    "https://docker.mirrors.sjtug.sjtu.edu.cn",
    "https://docker.mirrors.ustc.edu.cn",
    "https://mirror.iscas.ac.cn",
    "https://docker.rainbond.cc",
    "https://do.nark.eu.org",
    "https://dc.j8.work",
    "https://dockerproxy.com",
    "https://gst6rzl9.mirror.aliyuncs.com",
    "https://registry.docker-cn.com",
    "http://hub-mirror.c.163.com",
    "http://mirrors.ustc.edu.cn/",
    "https://mirrors.tuna.tsinghua.edu.cn/",
    "http://mirrors.sohu.com/"
  ],
  "insecure-registries": [
    "registry.docker-cn.com",
    "docker.mirrors.ustc.edu.cn"
  ],
  "debug": true,
  "experimental": false
}
EOF
        systemctl daemon-reload
        systemctl restart docker
    fi
    
    print_success "Docker 镜像源配置完成"
}

# 检查并安装pnpm
setup_pnpm() {
    print_info "检查并设置pnpm..."
    
    # 切换到前端目录
    cd frontend
    
    # 检查是否已安装pnpm
    if command -v pnpm &> /dev/null; then
        print_success "pnpm 已安装，版本: $(pnpm --version)"
    else
        print_warn "pnpm 未安装，正在安装..."
        
        # 方法1: 使用 npm 安装
        if command -v npm &> /dev/null; then
            print_info "使用npm安装pnpm..."
            npm install -g pnpm
        else
            # 方法2: 使用 curl 安装
            if command -v curl &> /dev/null; then
                print_info "使用curl安装pnpm..."
                curl -fsSL https://get.pnpm.io/install.sh | sh -
            else
                # 方法3: 使用 wget 安装
                if command -v wget &> /dev/null; then
                    print_info "使用wget安装pnpm..."
                    wget -qO- https://get.pnpm.io/install.sh | sh -
                else
                    print_error "无法安装pnpm: 没有找到可用的安装方法 (npm, curl, wget)"
                    cd ..
                    exit 1
                fi
            fi
        fi
        
        print_success "pnpm安装成功，版本: $(pnpm --version)"
    fi
    
    # 配置npm registry
    print_info "配置npm registry为淘宝镜像..."
    
    # 配置npm
    if command -v npm &> /dev/null; then
        npm config set registry https://registry.npmmirror.com
        print_success "npm registry配置完成"
    else
        print_warn "未找到npm命令，跳过npm registry配置"
    fi
    
    # 配置.npmrc文件
    echo "registry=https://registry.npmmirror.com" > .npmrc
    print_success ".npmrc文件配置完成"
    
    # 返回上级目录
    cd ..
    
    print_success "pnpm环境设置完成！"
}

# 默认参数
DEV_MODE=false
FORCE_BUILD=false
CONFIGURE_REGISTRY=false
PNPM_SETUP=false

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
        -r|--registry)
            CONFIGURE_REGISTRY=true
            shift
            ;;
        -p|--pnpm)
            PNPM_SETUP=true
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
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    print_error "未找到 Docker Compose，请先安装 Docker Compose"
    exit 1
fi

print_success "Docker 环境检查通过"

# 配置 Docker 镜像源（如果需要）
if [ "$CONFIGURE_REGISTRY" = true ]; then
    configure_docker_registry
fi

# 设置pnpm（如果需要）
if [ "$PNPM_SETUP" = true ]; then
    setup_pnpm
fi

# 检查环境变量文件
ENV_FILES=(".env" ".env.development" ".env.production" "backend/.env" "backend/.env.development" "backend/.env.production")
ENV_FILE_FOUND=false

for env_file in "${ENV_FILES[@]}"; do
    if [ -f "$env_file" ]; then
        print_success "找到环境变量文件: $env_file"
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
    $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE up $BUILD_OPTION -d
else
    $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE up $BUILD_OPTION -d
fi

# 等待服务启动
print_info "等待服务启动..."
sleep 5

# 检查服务状态
print_info "检查服务状态..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE ps

print_success "服务启动完成！"
print_success "后端服务地址: http://localhost:8080"
print_success "Redis服务地址: redis://localhost:6379"
# if [ "$DEV_MODE" = true ]; then
#     print_info "前端开发服务器地址: http://localhost:3000"
# else
#     print_info "前端服务地址: http://localhost"
# fi
print_success "监控指标地址: http://localhost:9100"

print_info "使用以下命令查看日志:"
print_info "  $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE logs -f"
print_info ""
print_info "使用以下命令停止服务:"
print_info "  $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down"