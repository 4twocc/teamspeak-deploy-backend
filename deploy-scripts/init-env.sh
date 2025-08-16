#!/bin/bash

# 初始化环境脚本
# 用于检查和安装必要的依赖项

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 打印带颜色的信息
print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} "
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

print_info "Initializing environment for TeamSpeak deployment..."

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]
  then print_warn "Please run as root"
  exit 1
fi

# 检查并安装 Docker (Ubuntu/Debian)
if ! command -v docker &> /dev/null
then
    print_warn "Docker not found. Installing Docker..."
    
    # 更新包索引
    apt-get update
    
    # 安装必要的包
    apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release \
        dpkg \
        gnupg \
    
    # 添加 Docker 官方 GPG 密钥
    curl -sSL https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/debian/gpg | gpg --dearmor > /usr/share/keyrings/docker-ce.gpg
    
    # 设置稳定版仓库
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-ce.gpg] https://mirrors.tuna.tsinghua.edu.cn/docker-ce/linux/debian \
      $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # 更新包索引
    apt-get update
    
    # 安装 Docker Engine
    apt-get install -y docker-ce docker-ce-cli containerd.io
    
    # 启动并启用 Docker 服务
    systemctl start docker
    systemctl enable docker
    
    print_success "Docker installed successfully"
else
    print_warn "Docker is already installed"
fi

# 验证 Docker 是否正常工作
if ! docker info &> /dev/null
then
    print_error "Error: Docker is not running. Starting Docker service..."
    systemctl start docker
    sleep 5
    
    if ! docker info &> /dev/null
    then
        print_error "Error: Failed to start Docker service"
        exit 1
    fi
fi

# 配置 Docker 镜像加速器和日志轮转
print_info "Configuring Docker daemon..."
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<EOF
{
    "registry-mirrors": [
        "https://docker.xuanyuan.me"
    ],
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "20m",
        "max-file": "3"
    },
    "userland-proxy": false,
    "ipv6": true,
    "fixed-cidr-v6": "fdb::/64",
    "experimental":true,
    "ip6tables":true
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker

print_success "Environment initialization completed successfully"
