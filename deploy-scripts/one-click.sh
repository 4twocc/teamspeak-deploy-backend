#!/bin/bash

set -e

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null
then
    echo "Error: Docker is not installed. Please install Docker first."
    exit 1
fi

# 检查 Docker 是否正在运行
if ! docker info &> /dev/null
then
    echo "Error: Docker is not running. Please start Docker daemon first."
    exit 1
fi

# 拉取 TeamSpeak 镜像
echo "Pulling TeamSpeak Docker image..."
docker pull teamspeak:latest

# 创建持久化数据目录
mkdir -p /var/lib/teamspeak/data

# 运行 TeamSpeak 容器
echo "Starting TeamSpeak server..."
docker run -d \
  --name teamspeak-main \
  -p 9987:9987/udp \
  -p 10011:10011/tcp \
  -p 30033:30033/tcp \
  -v /var/lib/teamspeak/data:/var/ts3server \
  -e TS3SERVER_LICENSE=accept \
  teamspeak:latest

echo "TeamSpeak server deployed successfully!"
echo "Server is running with container name: teamspeak-main"
echo "Data is persisted in: /var/lib/teamspeak/data"
echo "First run may take a minute to initialize. Check logs with: docker logs teamspeak-main"
echo "Admin token will be displayed in the logs on first run."
