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
        lsb-release
    
    # 添加 Docker 官方 GPG 密钥
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    
    # 设置稳定版仓库
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
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

echo "Environment initialization completed successfully"