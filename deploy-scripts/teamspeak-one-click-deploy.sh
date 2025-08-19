#!/bin/bash

# TeamSpeak一键部署脚本

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

print_info "TeamSpeak一键部署脚本"

# 检查必要的脚本文件是否存在
scripts=("init-env.sh" "open-ports.sh" "one-click.sh" "one-click-enhanced.sh")
for script in "${scripts[@]}"; do
    if [ ! -f "./$script" ]; then
        print_error "错误: 找不到脚本文件 $script"
        exit 1
    fi
done

# 确保脚本具有执行权限
print_info "检查并设置脚本执行权限..."
chmod +x ./init-env.sh ./open-ports.sh ./one-click.sh ./one-click-enhanced.sh 2>/dev/null || {
    print_warn "警告: 无法设置脚本执行权限，请确保以适当权限运行此脚本"
}

# 检查并初始化环境变量文件
check_env_files() {
    print_info "检查环境变量文件..."
    
    # 检查主目录.env文件
    if [ ! -f "../.env" ]; then
        if [ -f "../backend/.env.example" ]; then
            print_info "创建 .env 文件..."
            cp ../backend/.env.example ../.env
            print_success "已从 backend/.env.example 创建 .env"
            
            # 生成JWT_SECRET
            print_info "生成JWT_SECRET..."
            JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "your_generated_jwt_secret_here_replace_me")
            sed -i.bak "s|your_jwt_secret_key_here|$JWT_SECRET|g" ../.env
            rm -f ../.env.bak
            
            print_warn "请更新 .env 文件中的其他配置项，如 TEAMSPEAK_PASSWORD 等"
        else
            print_warn "未找到 backend/.env.example 文件，无法创建 .env"
        fi
    else
        print_success ".env 文件已存在"
    fi
    
    # 检查backend目录下的环境变量文件
    if [ ! -f "../backend/.env" ]; then
        if [ -f "../backend/.env.example" ]; then
            print_info "创建 backend/.env 文件..."
            cp ../backend/.env.example ../backend/.env
            print_success "已从 backend/.env.example 创建 backend/.env"
            
            # 生成JWT_SECRET
            print_info "生成JWT_SECRET..."
            JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "your_generated_jwt_secret_here_replace_me")
            sed -i.bak "s|your_jwt_secret_key_here|$JWT_SECRET|g" ../backend/.env
            rm -f ../backend/.env.bak
            
            print_warn "请更新 backend/.env 文件中的其他配置项，如 TEAMSPEAK_PASSWORD 等"
        else
            print_warn "未找到 backend/.env.example 文件，无法创建 backend/.env"
        fi
    else
        print_success "backend/.env 文件已存在"
    fi
}

while true; do
    print_info "=================================="
    print_info "请选择您要执行的操作："
    print_info "1. 初始化环境"
    print_info "2. 开启端口"
    print_info "3. 一键部署"
    print_info "4. 增强版一键部署（自动提取并保存凭证）"
    print_info "5. 执行所有步骤（推荐）"
    print_info "6. 执行所有步骤（增强版，推荐用于首次部署）"
    print_info "7. 退出"
    read -p "请输入您的选择（1-7）: " choice
    
    case $choice in 
        1)
            print_info "正在初始化环境..."
            if sudo ./init-env.sh; then
                print_success "环境初始化完成！"
            else
                print_error "环境初始化失败！"
            fi
            read -p "按回车键继续..."
            ;;
        2)
            print_info "正在开启端口..."
            if sudo ./open-ports.sh; then
                print_success "端口开启完成！"
            else
                print_error "端口开启失败！"
            fi
            read -p "按回车键继续..."
            ;;
        3)
            print_info "正在部署TeamSpeak服务..."
            if ./one-click.sh; then
                print_success "TeamSpeak部署完成！"
            else
                print_error "TeamSpeak部署失败！"
            fi
            read -p "按回车键继续..."
            ;;
        4)
            print_info "正在使用增强版部署TeamSpeak服务（自动提取并保存凭证）..."
            if ./one-click-enhanced.sh; then
                print_success "TeamSpeak部署完成，凭证已自动保存！"
            else
                print_error "TeamSpeak部署失败！"
            fi
            read -p "按回车键继续..."
            ;;
        5)
            print_info "开始执行所有部署步骤..."
            print_info "步骤1: 初始化环境"
            if sudo ./init-env.sh; then
                print_success "环境初始化完成！"
            else
                print_error "环境初始化失败！"
                read -p "按回车键继续..."
                continue
            fi
            
            print_info "步骤2: 开启端口"
            if sudo ./open-ports.sh; then
                print_success "端口开启完成！"
            else
                print_error "端口开启失败！"
                read -p "按回车键继续..."
                continue
            fi
            
            print_info "步骤3: 部署TeamSpeak"
            if ./one-click.sh; then
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
            if sudo ./init-env.sh; then
                print_success "环境初始化完成！"
            else
                print_error "环境初始化失败！"
                read -p "按回车键继续..."
                continue
            fi
            
            print_info "步骤2: 开启端口"
            if sudo ./open-ports.sh; then
                print_success "端口开启完成！"
            else
                print_error "端口开启失败！"
                read -p "按回车键继续..."
                continue
            fi
            
            print_info "步骤3: 增强版部署TeamSpeak（自动提取并保存凭证）"
            if ./one-click-enhanced.sh; then
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
            print_info "退出脚本"
            exit 0
            ;;
        *)
            print_warn "无效的选择，请输入1-7之间的数字"
            read -p "按回车键继续..."
            ;;
    esac
done