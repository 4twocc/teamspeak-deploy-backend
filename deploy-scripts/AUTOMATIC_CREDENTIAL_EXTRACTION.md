# TeamSpeak 凭证自动提取功能说明

## 概述

在首次部署 TeamSpeak 服务器时，系统会生成一些重要的凭证信息，如 ServerAdmin 密码、Server Query API Key 和 Server Admin Token。这些信息通常只在首次运行的日志中显示，之后就不再显示。

为了简化配置过程，我们提供了自动提取这些凭证并保存到 `.env` 文件的功能，这样后端服务就可以直接使用这些凭证，无需手动配置。

## 实现方式

我们提供了两种方式来自动提取 TeamSpeak 凭证：

### 1. 增强版部署脚本

增强版部署脚本 `one-click-enhanced.sh` 在部署 TeamSpeak 服务后会自动从日志中提取凭证并保存到 `.env` 文件中。

### 2. 独立的凭证提取工具

Go 程序 `extract_teamspeak_creds.go` 可以独立运行，从指定的日志文件中提取凭证并更新 `.env` 文件。

## 使用方法

### 方法一：使用增强版部署脚本（推荐）

1. 运行主部署脚本：
   ```bash
   cd deploy-scripts
   ./teamspeak-one-click-deploy.sh
   ```

2. 选择选项 6: "执行所有步骤（增强版，推荐用于首次部署）"

这将自动执行环境初始化、端口开放和增强版部署，部署完成后会自动提取并保存凭证。

### 方法二：手动运行增强版脚本

如果您已经完成了环境初始化和端口开放，可以直接运行增强版部署脚本：

```bash
cd deploy-scripts
./one-click-enhanced.sh
```

### 方法三：使用独立的凭证提取工具

如果您已经部署了 TeamSpeak 服务，但需要提取凭证，可以使用独立的工具：

```bash
cd backend
go run extract_teamspeak_creds.go /var/lib/teamspeak/data/first_run.log
```

## 凭证提取原理

### 日志分析

凭证提取工具会分析 TeamSpeak 容器的首次运行日志，查找以下模式的信息：

1. **ServerAdmin 密码**:
   ```
   serveradmin password= [PASSWORD]
   ```

2. **Server Query API Key**:
   ```
   API key: [API_KEY]
   ```

3. **Server Admin Token**:
   ```
   token=[TOKEN]
   ```

### .env 文件更新

提取到凭证后，工具会更新 `backend/.env` 文件，添加或更新以下环境变量：

- `TEAMSPEAK_PASSWORD` - ServerAdmin 密码
- `TEAMSPEAK_SERVER_QUERY_APIKEY` - Server Query API Key（如果存在）
- `TEAMSPEAK_SERVER_ADMIN_TOKEN` - Server Admin Token（如果存在）

如果 `.env` 文件已存在，工具会保留其他配置项，只更新 TeamSpeak 相关的凭证。

## 安全注意事项

1. **凭证保护**: 自动生成的 `.env` 文件包含敏感信息，请确保其权限设置正确，防止未授权访问：
   ```bash
   chmod 600 backend/.env
   ```

2. **备份**: 在更新 `.env` 文件之前，系统会自动创建备份文件，文件名为 `.env.backup.[timestamp]`。

3. **传输安全**: 在生产环境中，避免通过不安全的网络传输包含凭证的文件。

## 故障排除

### 凭证未提取成功

如果凭证提取失败，请检查：

1. TeamSpeak 容器是否正常运行：
   ```bash
   docker ps | grep teamspeak-main
   ```

2. 日志文件是否存在且包含凭证信息：
   ```bash
   cat /var/lib/teamspeak/data/first_run.log | grep -i password
   ```

3. 确保这是首次运行的日志，因为 TeamSpeak 在后续运行中不会再次显示凭证。

### .env 文件未更新

如果 `.env` 文件未按预期更新，请检查：

1. 是否有写入权限到 `backend/.env` 文件
2. 工具是否正确识别了项目根目录

## 最佳实践

1. **首次部署时使用**: 建议在首次部署 TeamSpeak 服务时使用增强版脚本，以自动完成凭证配置。

2. **定期检查**: 定期检查 `.env` 文件中的凭证是否仍然有效。

3. **版本控制**: 不要将包含真实凭证的 `.env` 文件提交到版本控制系统中。确保在 `.gitignore` 中添加 `.env` 文件：
   ```
   # Environment variables
   .env
   ```

4. **生产环境**: 在生产环境中，考虑使用更安全的凭证管理方案，如 HashiCorp Vault 或 AWS Secrets Manager。