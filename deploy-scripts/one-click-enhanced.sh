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
    
    # 检查日志文件是否存在
    if [ ! -f "$log_file" ]; then
        print_error "Log file does not exist: $log_file"
        return 1
    fi
    
    # 显示日志文件内容以便调试
    print_debug "Log file content:"
    while IFS= read -r line; do
        print_debug "$line"
    done < "$log_file"
    
    # 提取serveradmin密码 - 使用更广泛的匹配模式
    # TeamSpeak日志中密码可能出现在不同的格式中
    password=""
    
    # 尝试多种模式提取密码
    # 模式1: 查找"serveradmin password="格式
    password=$(grep -i "serveradmin password=" "$log_file" | sed -E 's/.*serveradmin password=([^ ]+).*/\1/' | head -1)
    
    # 模式2: 查找单独的password=模式
    if [ -z "$password" ]; then
        password=$(grep -E "password=[^ ]+" "$log_file" | grep -v "permissions\|database\|file\|config" | sed -E 's/.*password=([^ ]+).*/\1/' | head -1)
    fi
    
    # 模式3: 更宽松的匹配模式
    if [ -z "$password" ]; then
        password=$(grep -i "password" "$log_file" | grep -v "permissions\|database\|file\|config" | sed -E 's/.*[pP]assword[=: ]*([^ \t\n\r]+).*/\1/' | head -1)
    fi
    
    if [ -z "$password" ]; then
        print_error "Failed to extract password from log file"
        print_info "Please check the log file manually: $log_file"
        return 1
    fi
    
    print_info "Successfully extracted password"
    
    # 提取API key（如果存在）- 使用更准确的匹配
    api_key=$(grep -E 'API key:\s*\S+' "$log_file" | sed -E 's/.*API key:\s*(\S+).*/\1/' | head -1)
    
    # 提取token（如果存在）- 使用更准确的匹配
    token=$(grep -E 'token=[^\s]*' "$log_file" | sed -E 's/.*token=([^\s]+).*/\1/' | head -1)
    
    # 备份现有的.env文件（如果存在）
    if [ -f "$env_file" ]; then
        cp "$env_file" "$env_file.backup.$(date +%s)"
        print_info "Backed up existing .env file to $env_file.backup"
    fi
    
    # 读取现有的配置（如果.env文件存在）
    if [ -f "$env_file" ]; then
        # 读取非敏感配置行
        grep -v "^TEAMSPEAK_PASSWORD=\|^TEAMSPEAK_SERVER_QUERY_APIKEY=\|^TEAMSPEAK_SERVER_ADMIN_TOKEN=\|^JWT_SECRET=" "$env_file" > "$env_file.tmp"
    else
        # 如果.env文件不存在，尝试从.env.example复制
        env_example_file="$(dirname "$env_file")/.env.example"
        if [ -f "$env_example_file" ]; then
            cp "$env_example_file" "$env_file.tmp"
        else
            # 如果.env.example也不存在，创建空文件
            touch "$env_file.tmp"
        fi
    fi
    
    # 创建或更新.env文件
    {
        # 添加现有配置
        cat "$env_file.tmp"
        echo ""
        
        # 添加TeamSpeak凭证
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
    
    # 删除临时文件
    rm -f "$env_file.tmp"
    
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