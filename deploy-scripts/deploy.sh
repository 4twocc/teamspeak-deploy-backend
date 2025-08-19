#!/bin/bash

# TeamSpeak统一部署脚本
# 合并了环境初始化、端口开放、服务部署、凭证提取和清理功能

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

# 显示帮助信息
show_help() {
    echo "TeamSpeak 一键部署脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  init               初始化环境（安装Docker等）"
    echo "  ports              开放所需端口"
    echo "  deploy             部署TeamSpeak服务"
    echo "  deploy-enhanced    增强版部署（包含凭证提取）"
    echo "  cleanup            清理部署环境"
    echo "  extract-creds      从日志中提取凭证"
    echo "  all                执行所有步骤（推荐）"
    echo "  all-enhanced       执行所有步骤（增强版，首次部署推荐）"
    echo "  help               显示此帮助信息"
    echo ""
    echo "交互式使用:"
    echo "  不带参数运行脚本进入交互式菜单"
}

# 检查并初始化环境变量文件
check_env_files() {
    print_info "检查环境变量文件..."
    
    # 获取项目根目录路径
    project_root=$(dirname "$(dirname "$(readlink -f "$0")")")
    
    # 检查主目录.env文件
    if [ ! -f "$project_root/.env" ]; then
        if [ -f "$project_root/backend/.env.example" ]; then
            print_info "创建 .env 文件..."
            cp "$project_root/backend/.env.example" "$project_root/.env"
            print_success "已从 backend/.env.example 创建 .env"
            
            # 生成JWT_SECRET
            print_info "生成JWT_SECRET..."
            JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "your_generated_jwt_secret_here_replace_me")
            sed -i.bak "s|your_jwt_secret_key_here|$JWT_SECRET|g" "$project_root/.env"
            rm -f "$project_root/.env.bak"
            
            print_warn "请更新 .env 文件中的其他配置项，如 TEAMSPEAK_PASSWORD 等"
        else
            print_warn "未找到 backend/.env.example 文件，无法创建 .env"
        fi
    else
        print_success ".env 文件已存在"
    fi
    
    # 检查backend目录下的环境变量文件
    if [ ! -f "$project_root/backend/.env" ]; then
        if [ -f "$project_root/backend/.env.example" ]; then
            print_info "创建 backend/.env 文件..."
            cp "$project_root/backend/.env.example" "$project_root/backend/.env"
            print_success "已从 backend/.env.example 创建 backend/.env"
            
            # 生成JWT_SECRET
            print_info "生成JWT_SECRET..."
            JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "your_generated_jwt_secret_here_replace_me")
            sed -i.bak "s|your_jwt_secret_key_here|$JWT_SECRET|g" "$project_root/backend/.env"
            rm -f "$project_root/backend/.env.bak"
            
            print_warn "请更新 backend/.env 文件中的其他配置项，如 TEAMSPEAK_PASSWORD 等"
        else
            print_warn "未找到 backend/.env.example 文件，无法创建 backend/.env"
        fi
    else
        print_success "backend/.env 文件已存在"
    fi
}

# 初始化环境
init_environment() {
    print_info "初始化环境..."
    
    # 检查是否以root权限运行
    if [ "$EUID" -ne 0 ]
    then 
        print_warn "请以 root 身份运行"
        return 1
    fi

    # 检查并安装 Docker (Ubuntu/Debian)
    if ! command -v docker &> /dev/null
    then
        print_warn "Docker 未找到，正在安装 Docker..."
        
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
        
        print_success "Docker 安装成功"
    else
        print_info "Docker 已安装"
    fi

    # 验证 Docker 是否正常工作
    if ! docker info &> /dev/null
    then
        print_error "错误: Docker 未运行，正在启动 Docker 服务..."
        systemctl start docker
        sleep 5
        
        if ! docker info &> /dev/null
        then
            print_error "错误: 无法启动 Docker 服务"
            return 1
        fi
    fi

    # 配置 Docker 镜像加速器和日志轮转
    print_info "配置 Docker daemon..."
    mkdir -p /etc/docker
    tee /etc/docker/daemon.json <<EOF
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
    systemctl daemon-reload
    systemctl restart docker

    print_success "环境初始化完成"
    return 0
}

# 开放端口
open_ports() {
    print_info "打开 TeamSpeak 所需端口..."

    # 检查是否以root权限运行
    if [ "$EUID" -ne 0 ]
    then 
        print_warn "请以 root 身份运行"
        return 1
    fi

    # 检测系统类型
    if command -v ufw &> /dev/null; then
        # Ubuntu/Debian 使用 UFW
        print_info "检测到 UFW 防火墙，正在打开端口..."
        ufw allow 9987/udp comment 'TeamSpeak voice port'
        ufw allow 10011/tcp comment 'TeamSpeak server query port'
        ufw allow 30033/tcp comment 'TeamSpeak file transfer port'
        print_success "使用 UFW 成功打开端口"
    elif command -v firewall-cmd &> /dev/null; then
        # CentOS/RHEL 使用 firewalld
        print_info "检测到 firewalld，正在打开端口..."
        firewall-cmd --permanent --add-port=9987/udp --add-port=10011/tcp --add-port=30033/tcp
        firewall-cmd --reload
        print_success "使用 firewalld 成功打开端口"
    elif command -v iptables &> /dev/null; then
        # 检测是否为Debian系统
        if [ -f /etc/debian_version ]; then
            print_info "检测到 Debian 系统，使用 iptables 打开端口..."
            # 安装iptables-persistent（如果尚未安装）
            if ! dpkg -l | grep -q iptables-persistent; then
                print_info "正在安装 iptables-persistent..."
                apt-get update
                apt-get install -y iptables-persistent
            fi
        else
            print_info "检测到 iptables，正在打开端口..."
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
        print_warn "警告: 未检测到受支持的防火墙，请手动打开以下端口:"
        print_info "  - 9987/udp (Voice)"
        print_info "  - 10011/tcp (Server Query)"
        print_info "  - 30033/tcp (File Transfer)"
        return 1
    fi

    print_success "TeamSpeak 端口打开成功!"
    return 0
}

# 部署TeamSpeak服务
deploy_teamspeak() {
    print_info "部署 TeamSpeak 服务..."

    # 检查 Docker 是否安装
    if ! command -v docker &> /dev/null
    then
        print_error "错误: Docker 未安装，请先安装 Docker"
        return 1
    fi

    # 检查 Docker 是否正在运行
    if ! docker info &> /dev/null
    then
        print_error "错误: Docker 未运行，请先启动 Docker daemon"
        return 1
    fi

    # 拉取 TeamSpeak 镜像
    print_info "拉取 TeamSpeak Docker 镜像..."
    docker pull teamspeak:latest

    # 创建持久化数据目录
    mkdir -p /var/lib/teamspeak/data

    # 运行 TeamSpeak 容器
    print_info "启动 TeamSpeak 服务器..."
    docker run -d \
      --name teamspeak-main \
      -p 9987:9987/udp \
      -p 10011:10011/tcp \
      -p 30033:30033/tcp \
      -v /var/lib/teamspeak/data:/var/ts3server \
      -e TS3SERVER_LICENSE=accept \
      teamspeak:latest

    print_success "TeamSpeak 服务器部署成功!"
    print_success "服务器正在容器中运行: teamspeak-main"
    print_success "数据持久化存储在: /var/lib/teamspeak/data"

    # 等待一段时间让服务初始化
    print_info "等待服务器初始化..."
    sleep 10

    # 获取并保存首次日志
    print_info "保存首次运行日志..."
    first_log_file="/var/lib/teamspeak/data/first_run.log"
    docker logs teamspeak-main > "$first_log_file" 2>&1

    print_info "首次运行日志已保存到: $first_log_file"
    print_info "首次运行可能需要一些时间来初始化，日志已保存到文件"
    print_info "管理员令牌将显示在日志文件中"

    return 0
}

# 增强版部署（包含凭证提取）
deploy_teamspeak_enhanced() {
    print_info "执行增强版 TeamSpeak 部署..."

    # 检查 Docker 是否安装
    if ! command -v docker &> /dev/null
    then
        print_error "错误: Docker 未安装，请先安装 Docker"
        return 1
    fi

    # 检查 Docker 是否正在运行
    if ! docker info &> /dev/null
    then
        print_error "错误: Docker 未运行，请先启动 Docker daemon"
        return 1
    fi

    # 拉取 TeamSpeak 镜像
    print_info "拉取 TeamSpeak Docker 镜像..."
    docker pull teamspeak:latest

    # 创建持久化数据目录
    mkdir -p /var/lib/teamspeak/data

    # 运行 TeamSpeak 容器
    print_info "启动 TeamSpeak 服务器..."
    docker run -d \
      --name teamspeak-main \
      -p 9987:9987/udp \
      -p 10011:10011/tcp \
      -p 30033:30033/tcp \
      -v /var/lib/teamspeak/data:/var/ts3server \
      -e TS3SERVER_LICENSE=accept \
      teamspeak:latest

    print_success "TeamSpeak 服务器部署成功!"
    print_success "服务器正在容器中运行: teamspeak-main"
    print_success "数据持久化存储在: /var/lib/teamspeak/data"

    # 等待一段时间让服务初始化
    print_info "等待服务器初始化（约15秒）..."
    sleep 15

    # 获取并保存首次日志
    print_info "保存首次运行日志..."
    first_log_file="/var/lib/teamspeak/data/first_run.log"
    docker logs teamspeak-main > "$first_log_file" 2>&1

    print_info "首次运行日志已保存到: $first_log_file"

    # 从日志中提取敏感信息并写入.env文件
    print_info "从日志中提取敏感信息并生成 JWT 密钥..."

    # 获取项目根目录路径
    project_root=$(dirname "$(dirname "$(readlink -f "$0")")")
    env_file="$project_root/backend/.env"

    # 使用bash脚本提取凭证
    if extract_credentials "$first_log_file" "$env_file"; then
        print_info "凭证已保存到: $env_file"
        print_success "TeamSpeak 部署和凭证提取完成!"
        print_info "您现在可以使用自动配置的凭证运行后端服务"
        return 0
    else
        print_error "凭证提取失败!"
        return 1
    fi
}

# 检查并提取凭证的函数
extract_credentials() {
    local log_file="$1"
    local env_file="$2"
    
    # 使用bash脚本提取凭证
    print_info "使用 bash 脚本提取凭证..."
    
    # 检查日志文件是否存在
    if [ ! -f "$log_file" ]; then
        print_error "日志文件不存在: $log_file"
        return 1
    fi
    
    # 显示日志文件内容以便调试
    print_debug "日志文件内容:"
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
        password=$(grep -E "password=[^ ]+" "$log_file" | grep -v "permissions|database|file|config" | sed -E 's/.*password=([^ ]+).*/\1/' | head -1)
    fi
    
    # 模式3: 更宽松的匹配模式
    if [ -z "$password" ]; then
        password=$(grep -i "password" "$log_file" | grep -v "permissions|database|file|config" | sed -E 's/.*[pP]assword[=: ]*([^ \t\n\r]+).*/\1/' | head -1)
    fi
    
    if [ -z "$password" ]; then
        print_error "从日志文件中提取密码失败"
        print_info "请手动检查日志文件: $log_file"
        return 1
    fi
    
    print_info "成功提取密码"
    
    # 提取API key（如果存在）- 使用更准确的匹配
    api_key=$(grep -E 'API key:\s*\S+' "$log_file" | sed -E 's/.*API key:\s*(\S+).*/\1/' | head -1)
    
    # 提取token（如果存在）- 使用更准确的匹配
    token=$(grep -E 'token=[^\s]*' "$log_file" | sed -E 's/.*token=([^\s]+).*/\1/' | head -1)
    
    # 备份现有的.env文件（如果存在）
    if [ -f "$env_file" ]; then
        cp "$env_file" "$env_file.backup.$(date +%s)"
        print_info "备份现有 .env 文件到 $env_file.backup"
    fi
    
    # 读取现有的配置（如果.env文件存在）
    if [ -f "$env_file" ]; then
        # 读取非敏感配置行
        grep -v "^TEAMSPEAK_PASSWORD=|^TEAMSPEAK_SERVER_QUERY_APIKEY=|^TEAMSPEAK_SERVER_ADMIN_TOKEN=|^JWT_SECRET=" "$env_file" > "$env_file.tmp"
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
    
    print_success "成功更新 .env 文件中的凭证并生成 JWT 密钥!"
    return 0
}

# 清理环境
cleanup() {
    print_warn "清理 TeamSpeak 部署..."

    # 检查是否以root权限运行
    if [ "$EUID" -ne 0 ]
    then 
        print_warn "请以 root 身份运行"
        return 1
    fi

    # 停止并删除容器
    if docker ps -a --format '{{.Names}}' | grep -q '^teamspeak-main$'; then
        print_warn "停止 TeamSpeak 容器..."
        docker stop teamspeak-main || true
        print_warn "删除 TeamSpeak 容器..."
        docker rm teamspeak-main || true
    else
        print_info "未找到 TeamSpeak 容器"
    fi

    # 删除数据目录（可选，用户可以选择保留数据）
    print_warn "注意: 数据目录 /var/lib/teamspeak 已保留"
    print_success "要完全删除数据，请运行: rm -rf /var/lib/teamspeak"

    print_success "清理完成"
    return 0
}

# 交互式菜单
show_menu() {
    while true; do
        print_info "=================================="
        print_info "TeamSpeak 一键部署脚本"
        print_info "请选择您要执行的操作："
        print_info "1. 初始化环境"
        print_info "2. 开启端口"
        print_info "3. 部署TeamSpeak"
        print_info "4. 增强版部署（自动提取并保存凭证）"
        print_info "5. 执行所有步骤（推荐）"
        print_info "6. 执行所有步骤（增强版，首次部署推荐）"
        print_info "7. 提取凭证"
        print_info "8. 清理环境"
        print_info "9. 退出"
        read -p "请输入您的选择（1-9）: " choice
        
        case $choice in 
            1)
                print_info "正在初始化环境..."
                if init_environment; then
                    print_success "环境初始化完成！"
                else
                    print_error "环境初始化失败！"
                fi
                read -p "按回车键继续..."
                ;;
            2)
                print_info "正在开启端口..."
                if open_ports; then
                    print_success "端口开启完成！"
                else
                    print_error "端口开启失败！"
                fi
                read -p "按回车键继续..."
                ;;
            3)
                print_info "正在部署TeamSpeak服务..."
                if deploy_teamspeak; then
                    print_success "TeamSpeak部署完成！"
                else
                    print_error "TeamSpeak部署失败！"
                fi
                read -p "按回车键继续..."
                ;;
            4)
                print_info "正在使用增强版部署TeamSpeak服务（自动提取并保存凭证）..."
                if deploy_teamspeak_enhanced; then
                    print_success "TeamSpeak部署完成，凭证已自动保存！"
                else
                    print_error "TeamSpeak部署失败！"
                fi
                read -p "按回车键继续..."
                ;;
            5)
                print_info "开始执行所有部署步骤..."
                print_info "步骤1: 初始化环境"
                if ! init_environment; then
                    print_error "环境初始化失败！"
                    read -p "按回车键继续..."
                    continue
                fi
                print_success "环境初始化完成！"
                
                print_info "步骤2: 开启端口"
                if ! open_ports; then
                    print_error "端口开启失败！"
                    read -p "按回车键继续..."
                    continue
                fi
                print_success "端口开启完成！"
                
                print_info "步骤3: 部署TeamSpeak"
                if deploy_teamspeak; then
                    print_success "TeamSpeak部署完成！"
                    print_info "=================================="
                    print_success "部署已全部完成！"
                    print_info "首次运行日志已保存至 /var/lib/teamspeak/data/first_run.log"
                    print_info "您可以使用以下命令查看服务状态："
                    print_info "  docker ps | grep teamspeak"
                    print_info "  docker logs teamspeak-main"
                else
                    print_error "TeamSpeak部署失败！"
                fi
                read -p "按回车键继续..."
                ;;
            6)
                print_info "开始执行所有增强版部署步骤..."
                
                # 步骤0: 检查并初始化环境变量文件
                print_info "步骤0: 检查并初始化环境变量文件"
                check_env_files
                
                print_info "步骤1: 初始化环境"
                if ! init_environment; then
                    print_error "环境初始化失败！"
                    read -p "按回车键继续..."
                    continue
                fi
                print_success "环境初始化完成！"
                
                print_info "步骤2: 开启端口"
                if ! open_ports; then
                    print_error "端口开启失败！"
                    read -p "按回车键继续..."
                    continue
                fi
                print_success "端口开启完成！"
                
                print_info "步骤3: 增强版部署TeamSpeak（自动提取并保存凭证）"
                if deploy_teamspeak_enhanced; then
                    print_success "TeamSpeak部署完成，凭证已自动保存！"
                    print_info "=================================="
                    print_success "增强版部署已全部完成！"
                    print_info "首次运行日志已保存至 /var/lib/teamspeak/data/first_run.log"
                    print_info "TeamSpeak凭证已自动提取并保存到 backend/.env 文件"
                    print_info "您可以使用以下命令查看服务状态："
                    print_info "  docker ps | grep teamspeak"
                    print_info "  docker logs teamspeak-main"
                else
                    print_error "TeamSpeak部署失败！"
                fi
                read -p "按回车键继续..."
                ;;
            7)
                print_info "从日志中提取凭证..."
                # 获取项目根目录路径
                project_root=$(dirname "$(dirname "$(readlink -f "$0")")")
                env_file="$project_root/backend/.env"
                first_log_file="/var/lib/teamspeak/data/first_run.log"
                
                if extract_credentials "$first_log_file" "$env_file"; then
                    print_success "凭证提取完成！"
                else
                    print_error "凭证提取失败！"
                fi
                read -p "按回车键继续..."
                ;;
            8)
                print_info "正在清理环境..."
                if cleanup; then
                    print_success "环境清理完成！"
                else
                    print_error "环境清理失败！"
                fi
                read -p "按回车键继续..."
                ;;
            9)
                print_info "退出脚本"
                exit 0
                ;;
            *)
                print_warn "无效的选择，请输入1-9之间的数字"
                read -p "按回车键继续..."
                ;;
        esac
    done
}

# 主函数
main() {
    # 如果没有参数，显示交互式菜单
    if [ $# -eq 0 ]; then
        show_menu
        exit 0
    fi

    # 根据参数执行相应操作
    case "$1" in
        init)
            init_environment
            ;;
        ports)
            open_ports
            ;;
        deploy)
            deploy_teamspeak
            ;;
        deploy-enhanced)
            deploy_teamspeak_enhanced
            ;;
        extract-creds)
            # 获取项目根目录路径
            project_root=$(dirname "$(dirname "$(readlink -f "$0")")")
            env_file="$project_root/backend/.env"
            first_log_file="/var/lib/teamspeak/data/first_run.log"
            extract_credentials "$first_log_file" "$env_file"
            ;;
        cleanup)
            cleanup
            ;;
        all)
            init_environment && \
            open_ports && \
            deploy_teamspeak
            ;;
        all-enhanced)
            check_env_files && \
            init_environment && \
            open_ports && \
            deploy_teamspeak_enhanced
            ;;
        help|-h|--help)
            show_help
            ;;
        *)
            print_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数并传递所有参数
main "$@"