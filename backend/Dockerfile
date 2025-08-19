# backend/Dockerfile
FROM golang:1.24.3-alpine AS builder

WORKDIR /app
COPY . .

# 添加构建参数
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

# 使用国内镜像源替换默认源，解决 apk add 卡住的问题
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装gcc、musl-dev、sqlite-dev和upx以支持CGO和二进制压缩
RUN apk add --no-cache gcc musl-dev sqlite-dev upx

# 设置 CGO 环境变量以解决 SQLite3 在 Alpine Linux 上的编译问题
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV CGO_LDFLAGS="-lm -lsqlite3"
ENV CGO_CFLAGS="-DSQLITE_ENABLE_FTS5 -DSQLITE_ENABLE_RTREE -DSQLITE_ENABLE_JSON1 -DSQLITE_OMIT_DEPRECATED -D_LARGEFILE64_SOURCE"

# 启用CGO以支持SQLite并使用优化参数
# 使用 libsqlite3 标记强制使用系统 SQLite 库
RUN go build -ldflags="-s -w" -tags "libsqlite3" -o teamspeak-backend .

# 使用UPX压缩二进制文件
RUN upx --best --lzma teamspeak-backend

FROM alpine:latest AS production
# 使用国内镜像源替换默认源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# 从builder阶段复制二进制文件
COPY --from=builder /app/teamspeak-backend .

# 复制默认配置文件
COPY --from=builder /app/config.yaml ./config.yaml

# 创建数据目录
RUN mkdir -p ./data

EXPOSE 8080

# 使用非root用户运行
RUN adduser -D -s /bin/sh teamspeak
USER teamspeak

# 启动命令
CMD ["./teamspeak-backend"]