#!/bin/bash

# 增强版TeamSpeak一键部署脚本
# 此脚本会在部署完成后自动提取敏感信息并写入.env文件

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

# 检查并提取凭证的函数
extract_credentials() {
    local log_file="$1"
    local env_file="$2"
    
    # 使用bash脚本提取凭证
    print_info "Extracting credentials using bash script..."
    
    # 提取serveradmin密码
    password=$(grep -o 'password=\s*\S*' "$log_file" | head -1 | cut -d'=' -f2 | xargs)
    if [ -z "$password" ]; then
        print_error "Failed to extract password from log file"
        return 1
    fi
    
    # 提取API key（如果存在）
    api_key=$(grep -o 'API key:\s*\S*' "$log_file" | head -1 | cut -d':' -f2 | xargs)
    
    # 提取token（如果存在）
    token=$(grep -o 'token=[^\s]*' "$log_file" | head -1 | cut -d'=' -f2 | xargs)
    
    # 备份现有的.env文件（如果存在）
    if [ -f "$env_file" ]; then
        cp "$env_file" "$env_file.backup.$(date +%s)"
        print_info "Backed up existing .env file to $env_file.backup"
    fi
    
    # 创建或更新.env文件
    {
        echo "# TeamSpeak Credentials (auto-generated)"
        echo "TEAMSPEAK_SERVER_ADMIN_USERNAME=serveradmin"
        echo "TEAMSPEAK_PASSWORD=$password"
        if [ -n "$api_key" ]; then
            echo "TEAMSPEAK_SERVER_QUERY_APIKEY=$api_key"
        fi
        if [ -n "$token" ]; then
            echo "TEAMSPEAK_SERVER_ADMIN_TOKEN=$token"
        fi
        echo ""
        echo "# JWT Secret (auto-generated)"
        # 生成JWT密钥
        if command -v openssl &> /dev/null; then
            jwt_secret=$(openssl rand -base64 32)
            echo "JWT_SECRET=$jwt_secret"
        else
            # 如果没有openssl，生成一个简单的密钥
            jwt_secret=$(date | md5sum | cut -d' ' -f1)
            echo "JWT_SECRET=$jwt_secret"
        fi
    } > "$env_file"
    
    print_success "Successfully updated .env file with credentials and generated JWT secret!"
    return 0
}

set -e

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null
then
    print_error "Error: Docker is not installed. Please install Docker first."
    exit 1
fi

# 检查 Docker 是否正在运行
if ! docker info &> /dev/null
then
    print_error "Error: Docker is not running. Please start Docker daemon first."
    exit 1
fi

# 拉取 TeamSpeak 镜像
print_info "Pulling TeamSpeak Docker image..."
docker pull teamspeak:latest

# 创建持久化数据目录
mkdir -p /var/lib/teamspeak/data

# 运行 TeamSpeak 容器
print_info "Starting TeamSpeak server..."
docker run -d \
  --name teamspeak-main \
  -p 9987:9987/udp \
  -p 10011:10011/tcp \
  -p 30033:30033/tcp \
  -v /var/lib/teamspeak/data:/var/ts3server \
  -e TS3SERVER_LICENSE=accept \
  teamspeak:latest

print_success "TeamSpeak server deployed successfully!"
print_success "Server is running with container name: teamspeak-main"
print_success "Data is persisted in: /var/lib/teamspeak/data"

# 等待一段时间让服务初始化
print_info "Waiting for server to initialize..."
sleep 15

# 获取并保存首次日志
print_info "Saving first log to file..."
first_log_file="/var/lib/teamspeak/data/first_run.log"
docker logs teamspeak-main > "$first_log_file" 2>&1

print_info "First run log has been saved to: $first_log_file"

# 从日志中提取敏感信息并写入.env文件
print_info "Extracting sensitive information from logs and generating JWT secret..."

# 获取项目根目录路径
project_root=$(dirname "$(dirname "$(readlink -f "$0")")")
env_file="$project_root/backend/.env"

# 使用bash脚本提取凭证（替代之前的Go程序）
if extract_credentials "$first_log_file" "$env_file"; then
    print_info "Credentials saved to: $env_file"
    print_success "TeamSpeak deployment and credential extraction completed successfully!"
    print_info "You can now use the backend service with the automatically configured credentials."
else
    print_error "Failed to extract credentials!"
    exit 1
fi