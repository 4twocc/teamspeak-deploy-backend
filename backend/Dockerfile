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

# 使用非root用户运行
RUN adduser -D -s /bin/sh teamspeak
# 设置数据目录的所有权
RUN chown teamspeak:teamspeak ./data

USER teamspeak

# 启动命令
CMD ["./teamspeak-backend"]