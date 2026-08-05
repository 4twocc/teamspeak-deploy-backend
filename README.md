# TeamSpeak One-Click Deploy

一个用于一键部署TeamSpeak 服务的工具。

## 功能特性

- 一键部署 TeamSpeak 服务

## 安装和部署

### 环统要求

- Debian/Ubuntu 系统（推荐 Ubuntu 20.04+）
- Docker 和 Docker Compose
- 至少 1GB 内存

## 部署脚本

项目包含以下部署脚本：

```bash
deploy-scripts/
├── deploy.sh                 # 统一部署脚本（推荐使用）
├── init-env.sh              # 环境初始化脚本
├── open-ports.sh            # 端口开放脚本
├── cleanup.sh               # 清理脚本
└── AUTOMATIC_CREDENTIAL_EXTRACTION.md # 凭证自动提取说明文档
```

### 使用方法

推荐使用统一的部署脚本：

```bash
# 交互式使用
cd deploy-scripts
./deploy.sh

# 命令行使用
cd deploy-scripts
./deploy.sh all-enhanced  # 执行完整的增强部署（包括凭证提取）
```

### 脚本功能说明

1. **deploy.sh**: 统一部署脚本，包含以下子命令：
   - `init`: 初始化环境（安装Docker等）
   - `ports`: 开放所需端口
   - `deploy`: 部署TeamSpeak服务
   - `deploy-enhanced`: 增强版部署（包含凭证提取）
   - `cleanup`: 清理部署环境
   - `extract-creds`: 从日志中提取凭证
   - `all`: 执行所有基本步骤
   - `all-enhanced`: 执行所有增强步骤（推荐用于首次部署）

2. **init-env.sh**: 环境初始化脚本，用于安装Docker等依赖

3. **open-ports.sh**: 端口开放脚本，用于开放TeamSpeak所需端口

4. **cleanup.sh**: 清理脚本，用于停止和删除TeamSpeak服务

## 开发指南

### 后端开发

1. 安装 Go 1.24.3+
2. 安装依赖: `go mod tidy`
3. 运行服务: `go run main.go`

## 常见问题

### Docker镜像拉取失败

如果遇到以下错误：
```
Get "https://registry-1.docker.io/v2/": net/http: request canceled while waiting for connection
```

这通常是因为网络连接问题导致无法访问Docker Hub。可以通过以下方式解决：

1. 使用国内Docker镜像源：
   ```bash
   # 运行项目提供的修复脚本（需要sudo权限）
   sudo ./fix-docker-registry.sh
   ```

2. 或者手动配置Docker镜像源：
   ```bash
   sudo mkdir -p /etc/docker
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
      ]
   }
   EOF
   sudo systemctl daemon-reload
   sudo systemctl restart docker
   ```

### 凭证提取失败

如果在部署TeamSpeak服务器后无法自动提取凭证，请检查：
1. TeamSpeak容器是否正常运行：`docker ps | grep teamspeak`
2. 日志文件是否存在：`ls /var/lib/teamspeak/data/first_run.log`
3. 日志文件中是否包含凭证信息：`cat /var/lib/teamspeak/data/first_run.log | grep -i password`

