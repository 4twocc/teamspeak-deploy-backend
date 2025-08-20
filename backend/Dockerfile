# backend/Dockerfile
FROM golang:1.24.3-alpine AS builder

WORKDIR /app
COPY . .

# 添加构建参数
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

# 使用国内镜像源替换默认源，解决 apk add 卡住的问题
RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories

# 安装gcc、musl-dev、sqlite-dev和upx以支持CGO和二进制压缩
RUN apk add --no-cache gcc g++ musl-dev make sqlite sqlite-dev upx

# 使用 go-sqlite3 推荐的 build tag，避免 load_extension 问题
RUN CGO_ENABLED=1 GOOS=linux go build -tags 'sqlite_omit_load_extension libsqlite3' -ldflags="-s -w" -o teamspeak-backend .

# 使用UPX压缩二进制文件
RUN upx --best --lzma teamspeak-backend

FROM alpine:latest AS production
# 使用国内镜像源替换默认源
RUN sed -i 's#https\?://dl-cdn.alpinelinux.org/alpine#https://mirrors.tuna.tsinghua.edu.cn/alpine#g' /etc/apk/repositories
RUN apk --no-cache add sqlite-libs ca-certificates tzdata
WORKDIR /app

# 从builder阶段复制二进制文件
COPY --from=builder /app/teamspeak-backend .

# 复制默认配置文件
COPY --from=builder /app/config.yaml ./config.yaml

# 创建数据目录并设置适当的权限和所有权
RUN mkdir -p ./data && \
    chmod 777 ./data

EXPOSE 8080
# 安装 su-exec 用于在容器中安全降权
RUN apk add --no-cache su-exec

# 创建非 root 用户 teamspeak（entrypoint 将处理 chown 并降权运行）
RUN adduser -D -s /bin/sh teamspeak

# 复制 entrypoint 脚本并赋予执行权限
COPY --chown=root:root ./entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# 确保工作目录权限正确
RUN chown -R teamspeak:teamspeak /app && \
    chmod -R 777 /app/data

# ENTRYPOINT 以 root 身份运行，脚本将修复挂载目录权限并使用 su-exec 切换到 teamspeak
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# 默认命令参数 — 后端二进制
CMD ["./teamspeak-backend"]
