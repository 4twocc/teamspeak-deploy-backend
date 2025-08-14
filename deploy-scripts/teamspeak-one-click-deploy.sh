#!/bin/bash

# TeamSpeak一键部署脚本

echo "TeamSpeak一键部署脚本"

# 检查必要的脚本文件是否存在
scripts=("init-env.sh" "open-ports.sh" "one-click.sh")
for script in "${scripts[@]}"; do
    if [ ! -f "./$script" ]; then
        echo "错误: 找不到脚本文件 $script"
        exit 1
    fi
done

# 确保脚本具有执行权限
echo "检查并设置脚本执行权限..."
chmod +x ./init-env.sh ./open-ports.sh ./one-click.sh 2>/dev/null || {
    echo "警告: 无法设置脚本执行权限，请确保以适当权限运行此脚本"
}

while true; do
    echo "=================================="
    echo "请选择您要执行的操作："
    echo "1. 初始化环境"
    echo "2. 开启端口"
    echo "3. 一键部署"
    echo "4. 执行所有步骤（推荐）"
    echo "5. 退出"
    read -p "请输入您的选择（1-5）: " choice
    
    case $choice in 
        1)
            echo "正在初始化环境..."
            if sudo ./init-env.sh; then
                echo "环境初始化完成！"
            else
                echo "环境初始化失败！"
            fi
            read -p "按回车键继续..."
            ;;
        2)
            echo "正在开启端口..."
            if sudo ./open-ports.sh; then
                echo "端口开启完成！"
            else
                echo "端口开启失败！"
            fi
            read -p "按回车键继续..."
            ;;
        3)
            echo "正在部署TeamSpeak服务..."
            if ./one-click.sh; then
                echo "TeamSpeak部署完成！"
            else
                echo "TeamSpeak部署失败！"
            fi
            read -p "按回车键继续..."
            ;;
        4)
            echo "开始执行所有部署步骤..."
            echo "步骤1: 初始化环境"
            if sudo ./init-env.sh; then
                echo "环境初始化完成！"
            else
                echo "环境初始化失败！"
                read -p "按回车键继续..."
                continue
            fi
            
            echo "步骤2: 开启端口"
            if sudo ./open-ports.sh; then
                echo "端口开启完成！"
            else
                echo "端口开启失败！"
                read -p "按回车键继续..."
                continue
            fi
            
            echo "步骤3: 部署TeamSpeak"
            if ./one-click.sh; then
                echo "TeamSpeak部署完成！"
                echo "=================================="
                echo "部署已全部完成！"
                echo "您可以使用以下命令查看服务状态："
                echo "  docker ps | grep teamspeak"
                echo "  docker logs teamspeak-main"
            else
                echo "TeamSpeak部署失败！"
            fi
            read -p "按回车键继续..."
            ;;
        5)
            echo "退出脚本"
            exit 0
            ;;
        *)
            echo "无效的选择，请输入1-5之间的数字"
            read -p "按回车键继续..."
            ;;
    esac
done
