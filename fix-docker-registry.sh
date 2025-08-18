#!/bin/bash

# 修复Docker镜像源连接问题的脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 检查是否以root权限运行
check_root() {
    if [ "$EUID" -ne 0 ] && ! command -v sudo &> /dev/null; then
        print_error "此脚本需要root权限，请使用sudo运行或以root用户身份运行"
        exit 1
    fi
}

# 配置Docker镜像源
configure_docker_registry() {
    print_info "配置Docker镜像源以解决连接问题..."
    
    # 创建Docker配置目录
    if command -v sudo &> /dev/null; then
        sudo mkdir -p /etc/docker
    else
        mkdir -p /etc/docker
    fi
    
    # 配置国内镜像源
    print_info "配置国内镜像源..."
    
    if command -v sudo &> /dev/null; then
        sudo tee /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://mirror.aliyuncs.com",
    "https://docker.m.daocloud.io",
    "https://docker.nju.edu.cn",
    "https://docker.mirrors.ustc.edu.cn",
    "https://mirror.iscas.ac.cn",
    "https://dockerproxy.com",
    "https://hub-mirror.c.163.com"
  ],
  "insecure-registries": [],
  "debug": true
}
EOF
    else
        tee /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://mirror.aliyuncs.com",
    "https://docker.m.daocloud.io",
    "https://docker.nju.edu.cn",
    "https://docker.mirrors.ustc.edu.cn",
    "https://mirror.iscas.ac.cn",
    "https://dockerproxy.com",
    "https://hub-mirror.c.163.com"
  ],
  "insecure-registries": [],
  "debug": true
}
EOF
    fi
    
    print_success "Docker镜像源配置完成"
}

# 重启Docker服务
restart_docker() {
    print_info "重启Docker服务以应用配置..."
    
    if command -v sudo &> /dev/null; then
        if sudo systemctl daemon-reload; then
            print_info "重新加载systemd配置成功"
        else
            print_warn "重新加载systemd配置失败"
        fi
        
        if sudo systemctl restart docker; then
            print_success "Docker服务重启成功"
        else
            print_error "Docker服务重启失败"
            exit 1
        fi
    else
        if systemctl daemon-reload; then
            print_info "重新加载systemd配置成功"
        else
            print_warn "重新加载systemd配置失败"
        fi
        
        if systemctl restart docker; then
            print_success "Docker服务重启成功"
        else
            print_error "Docker服务重启失败"
            exit 1
        fi
    fi
}

# 测试Docker连接
test_docker_connection() {
    print_info "测试Docker连接..."
    
    # 等待Docker服务完全启动
    sleep 5
    
    # 尝试拉取一个小型镜像来测试连接
    if docker pull hello-world:latest &> /dev/null; then
        print_success "Docker连接测试成功"
        # 清理测试镜像
        docker rmi hello-world:latest &> /dev/null || true
        return 0
    else
        print_error "Docker连接测试失败"
        return 1
    fi
}

# 主函数
main() {
    print_info "开始修复Docker镜像源连接问题"
    
    # 检查权限
    check_root
    
    # 配置Docker镜像源
    configure_docker_registry
    
    # 重启Docker服务
    restart_docker
    
    # 测试连接
    if test_docker_connection; then
        print_success "Docker镜像源连接问题修复完成！"
        print_info "现在可以正常拉取Docker镜像了"
    else
        print_error "Docker连接测试仍然失败，请检查网络连接或尝试其他解决方案"
        exit 1
    fi
}

# 显示使用说明
show_usage() {
    print_info "使用方法:"
    print_info "  sudo ./fix-docker-registry.sh"
    print_info ""
    print_info "此脚本将:"
    print_info "  1. 配置国内Docker镜像源"
    print_info "  2. 重启Docker服务"
    print_info "  3. 测试Docker连接"
}

# 解析命令行参数
if [[ $# -gt 0 ]]; then
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "未知选项: $1"
            show_usage
            exit 1
            ;;
    esac
fi

# 运行主函数
main