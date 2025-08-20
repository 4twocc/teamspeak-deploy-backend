# TeamSpeak One-Click Deploy

一个用于一键部署和管理 TeamSpeak 服务器的工具。

## 功能特性

- 一键部署 TeamSpeak 服务器
- Web 管理界面
- 实时监控和告警
- 实例管理
- 用户认证和权限管理

## 安装和部署

### 环统要求

- Debian/Ubuntu 系统（推荐 Ubuntu 20.04+）
- Docker 和 Docker Compose
- 至少 1GB 内存

### 部署步骤

1. 克隆项目：
   ```bash
   git clone <repository-url>
   cd teamspeak-one-click-deploy
   ```

2. 配置环境变量：
   ```bash
   # 生产环境
   cp backend/.env.example backend/.env
   # 开发环境
   cp backend/.env.development.example backend/.env.development
   # 编辑相应的环境变量文件，设置敏感配置
   ```

3. 启动服务：
   ```bash
   # 生产环境
   docker-compose up -d
   # 开发环境
   docker-compose -f docker-compose.dev.yml up -d
   ```

4. 获取管理员凭证：
   ```bash
   # 查看TeamSpeak服务器管理员凭证
   docker exec teamspeak_container_id cat /var/lib/teamspeak/data/first_run.log | grep -i password
   ```

4. 访问 Web 界面：
   - API: http://localhost:8080

## 安全配置

项目使用环境变量来管理敏感信息，避免将密码等敏感数据提交到代码仓库。

### 敏感配置项

1. **TeamSpeak ServerQuery 密码**
   - 环境变量: `TEAMSPEAK_PASSWORD`
   - 用途: 连接到 TeamSpeak 服务器的查询接口，用于监控和管理

2. **JWT Secret**
   - 环境变量: `JWT_SECRET`
   - 用途: 签名和验证用户认证 token
   - 生成安全密钥: `openssl rand -base64 32`

3. **TeamSpeak Server Query API Key**
   - 环境变量: `TEAMSPEAK_SERVER_QUERY_APIKEY`
   - 用途: 用于访问TeamSpeak Server Query API的API密钥

4. **TeamSpeak Server Admin Token**
   - 环境变量: `TEAMSPEAK_SERVER_ADMIN_TOKEN`
   - 用途: 用于获取管理员权限的令牌

5. **TeamSpeak Server Admin Username**
   - 环境变量: `TEAMSPEAK_USERNAME`
   - 用途: TeamSpeak服务器管理员用户名

### 配置方法

1. 复制示例环境变量文件：
   ```bash
   # 生产环境
   cp backend/.env.example backend/.env
   # 开发环境
   cp backend/.env.development.example backend/.env.development
   ```

2. 编辑相应的环境变量文件，设置实际值：
   ```bash
   TEAMSPEAK_PASSWORD=your_actual_teamspeak_password
   JWT_SECRET=your_actual_jwt_secret
   TEAMSPEAK_SERVER_QUERY_APIKEY=your_actual_api_key
   TEAMSPEAK_SERVER_ADMIN_TOKEN=your_actual_admin_token
   TEAMSPEAK_USERNAME=serveradmin
   ```

3. 生成安全的JWT密钥：
   ```bash
   openssl rand -base64 32
   ```

## API 文档

API 文档可通过 `/api/docs` 端点访问。

## 目录结构

```
.
├── backend             # 后端服务
│   ├── api             # API 路由
│   ├── auth            # 认证模块
│   ├── config          # 配置管理
│   ├── database        # 数据库访问
│   ├── deploy          # 部署模块
│   ├── instance        # 实例管理
│   ├── monitor         # 监控模块
│   ├── users           # 用户管理
│   ├── utils           # 工具函数
│   ├── config.yaml     # 配置文件
│   ├── entrypoint.sh   # TeamSpeak 容器入口点脚本
│   └── main.go         # 主程序入口
│ 
├── deploy-scripts      # 部署脚本
└── docker-compose.yml
```

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

