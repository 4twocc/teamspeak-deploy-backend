#!/bin/bash

# 自动开启TeamSpeak所需的端口
# TeamSpeak默认使用以下端口:
#  - 9987/udp: 语音端口
#  - 10011/tcp: 服务端查询端口
#  - 30033/tcp: 文件传输端口

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

set -e

print_info "打开 TeamSpeak 所需端口..."

# 检查是否以root权限运行
if [ "$EUID" -ne 0 ]
  then print_warn "请以 root 身份运行"
  exit 1
fi

# 检测系统类型
if command -v ufw &> /dev/null; then
    # Ubuntu/Debian 使用 UFW
    print_warn "检测到 UFW 防火墙，正在打开端口..."
    ufw allow 9987/udp comment 'TeamSpeak voice port'
    ufw allow 10011/tcp comment 'TeamSpeak server query port'
    ufw allow 30033/tcp comment 'TeamSpeak file transfer port'
    print_success "使用 UFW 成功打开端口"
elif command -v firewall-cmd &> /dev/null; then
    # CentOS/RHEL 使用 firewalld
    print_warn "检测到 firewalld，正在打开端口..."
    firewall-cmd --permanent --add-port=9987/udp --add-port=10011/tcp --add-port=30033/tcp
    firewall-cmd --reload
    print_success "使用 firewalld 成功打开端口"
elif command -v iptables &> /dev/null; then
    # 检测是否为Debian系统
    if [ -f /etc/debian_version ]; then
        print_warn "检测到 Debian 系统，使用 iptables 打开端口..."
        # 安装iptables-persistent（如果尚未安装）
        if ! dpkg -l | grep -q iptables-persistent; then
            print_info "正在安装 iptables-persistent..."
            apt-get update
            apt-get install -y iptables-persistent
        fi
    else
        print_warn "检测到 iptables，正在打开端口..."
    fi
    
    # 添加iptables规则
    iptables -A INPUT -p udp --dport 9987 -j ACCEPT
    iptables -A INPUT -p tcp --dport 10011 -j ACCEPT
    iptables -A INPUT -p tcp --dport 30033 -j ACCEPT
    
    # 保存规则
    if [ -f /etc/debian_version ]; then
        # Debian系统使用netfilter-persistent保存规则
        print_info "为 Debian 保存 iptables 规则..."
        netfilter-persistent save
    elif command -v netfilter-persistent &> /dev/null; then
        # 使用netfilter-persistent（如果可用）
        print_info "使用 netfilter-persistent 保存 iptables 规则..."
        netfilter-persistent save
    else
        # 手动保存规则（通用方法）
        print_info "手动保存 iptables 规则..."
        # 检查目标目录是否存在，如果不存在则创建
        if [ ! -d "/etc/iptables" ]; then
            mkdir -p /etc/iptables
        fi
        iptables-save > /etc/iptables/rules.v4 2>/dev/null || echo "Warning: Unable to save iptables rules. They may be lost after reboot."
    fi
    
    print_success "使用 iptables 成功打开端口"
else
    print_warn "Warning: No supported firewall detected. Please manually open the following ports:"
    print_info "  - 9987/udp (Voice)"
    print_info "  - 10011/tcp (Server Query)"
    print_info "  - 30033/tcp (File Transfer)"
    exit 1
fi

print_success "TeamSpeak ports opened successfully!"
print_success "You can now deploy TeamSpeak service using docker compose."