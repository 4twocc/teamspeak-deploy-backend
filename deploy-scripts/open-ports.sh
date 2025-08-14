#!/bin/bash

# 自动开启TeamSpeak所需的端口
# TeamSpeak默认使用以下端口:
#  - 9987/udp: 语音端口
#  - 10011/tcp: 服务端查询端口
#  - 30033/tcp: 文件传输端口

set -e

echo "Opening TeamSpeak required ports..."

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]
  then echo "Please run as root"
  exit 1
fi

# 检测系统类型
if command -v ufw &> /dev/null; then
    # Ubuntu/Debian 使用 UFW
    echo "Detected UFW firewall. Opening ports..."
    ufw allow 9987/udp comment 'TeamSpeak voice port'
    ufw allow 10011/tcp comment 'TeamSpeak server query port'
    ufw allow 30033/tcp comment 'TeamSpeak file transfer port'
    echo "Ports opened successfully with UFW"
elif command -v firewall-cmd &> /dev/null; then
    # CentOS/RHEL 使用 firewalld
    echo "Detected firewalld. Opening ports..."
    firewall-cmd --permanent --add-port=9987/udp --add-port=10011/tcp --add-port=30033/tcp
    firewall-cmd --reload
    echo "Ports opened successfully with firewalld"
elif command -v iptables &> /dev/null; then
    # 检测是否为Debian系统
    if [ -f /etc/debian_version ]; then
        echo "Detected Debian system. Opening ports with iptables..."
        # 安装iptables-persistent（如果尚未安装）
        if ! dpkg -l | grep -q iptables-persistent; then
            echo "Installing iptables-persistent..."
            apt-get update
            apt-get install -y iptables-persistent
        fi
    else
        echo "Detected iptables. Opening ports..."
    fi
    
    # 添加iptables规则
    iptables -A INPUT -p udp --dport 9987 -j ACCEPT
    iptables -A INPUT -p tcp --dport 10011 -j ACCEPT
    iptables -A INPUT -p tcp --dport 30033 -j ACCEPT
    
    # 保存规则
    if [ -f /etc/debian_version ]; then
        # Debian系统使用netfilter-persistent保存规则
        echo "Saving iptables rules for Debian..."
        netfilter-persistent save
    elif command -v service &> /dev/null && service iptables save 2>/dev/null; then
        # 尝试使用service命令保存（仅在支持的系统上）
        echo "Saving iptables rules with service command..."
    elif command -v netfilter-persistent &> /dev/null; then
        # 使用netfilter-persistent（如果可用）
        echo "Saving iptables rules with netfilter-persistent..."
        netfilter-persistent save
    else
        # 手动保存规则（通用方法）
        echo "Saving iptables rules manually..."
        iptables-save > /etc/iptables/rules.v4 2>/dev/null || echo "Warning: Unable to save iptables rules. They may be lost after reboot."
    fi
    
    echo "Ports opened successfully with iptables"
else
    echo "Warning: No supported firewall detected. Please manually open the following ports:"
    echo "  - 9987/udp (Voice)"
    echo "  - 10011/tcp (Server Query)"
    echo "  - 30033/tcp (File Transfer)"
    exit 1
fi

echo "TeamSpeak ports opened successfully!"
echo "You can now deploy TeamSpeak service."
