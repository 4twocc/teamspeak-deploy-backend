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

4. 访问 Web 界面：
   - 前端: http://localhost
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
   ```

## API 文档

API 文档可通过 `/api/docs` 端点访问。

## 目录结构

```
.
├── backend          # 后端服务
│   ├── api          # API 路由
│   ├── auth         # 认证模块
│   ├── config       # 配置管理
│   ├── database     # 数据库访问
│   ├── deploy       # 部署模块
│   ├── instance     # 实例管理
│   ├── monitor      # 监控模块
│   ├── users        # 用户管理
│   ├── utils        # 工具函数
│   ├── config.yaml  # 配置文件
│   └── main.go      # 主程序入口
├── deploy-scripts   # 部署脚本
├── frontend         # 前端界面
└── docker-compose.yml
```

## 开发指南

### 后端开发

1. 安装 Go 1.24+
2. 安装依赖: `go mod tidy`
3. 运行服务: `go run main.go`

### 前端开发

1. 安装 Node.js 22+
2. 安装依赖: `npm install`
3. 启动开发服务器: `npm run dev`

## 贡献指南
1. Fork 本仓库
2. 创建一个分支: `git checkout -b feature/xxx`
3. 提交你的修改: `git commit -m 'feat: add xxx'`
4. 推送到远程分支: `git push origin feature/xxx`
5. 创建一个 Pull Request
6. 等待审核
7. 合并 Pull Request
8. 恭喜，你已贡献了代码！