#!/bin/bash

# 初始化环境脚本
# 用于检查和安装必要的依赖项

set -e

echo "Initializing environment for TeamSpeak deployment..."

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit 1
fi

# 检查并安装 Docker (Ubuntu/Debian)
if ! command -v docker &> /dev/null
then
    echo "Docker not found. Installing Docker..."
    
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
    
    echo "Docker installed successfully"
else
    echo "Docker is already installed"
fi

# 验证 Docker 是否正常工作
if ! docker info &> /dev/null
then
    echo "Error: Docker is not running. Starting Docker service..."
    systemctl start docker
    sleep 5
    
    if ! docker info &> /dev/null
    then
        echo "Error: Failed to start Docker service"
        exit 1
    fi
fi

# 配置 Docker 镜像加速器和日志轮转
echo "Configuring Docker daemon..."
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

echo "Environment initialization completed successfully"
