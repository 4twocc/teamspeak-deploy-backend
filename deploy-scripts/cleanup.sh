#!/bin/bash

# 清理脚本
# 用于停止和删除TeamSpeak服务及相关数据

set -e

echo "Cleaning up TeamSpeak deployment..."

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit 1
fi

# 停止并删除容器
if docker ps -a --format '{{.Names}}' | grep -q '^teamspeak-main$'; then
    echo "Stopping TeamSpeak container..."
    docker stop teamspeak-main || true
    echo "Removing TeamSpeak container..."
    docker rm teamspeak-main || true
else
    echo "No TeamSpeak container found"
fi

# 删除数据目录（可选，用户可以选择保留数据）
echo "Note: Data directory /var/lib/teamspeak is preserved."
echo "To remove it completely, run: rm -rf /var/lib/teamspeak"

echo "Cleanup completed successfully"