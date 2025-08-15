# backend/Dockerfile
FROM golang:1.24.3-alpine AS builder

WORKDIR /app
COPY . .

# 添加构建参数
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o teamspeak-backend .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 从builder阶段复制二进制文件
COPY --from=builder /app/teamspeak-backend .

# 复制默认配置文件
COPY --from=builder /app/config.yaml ./config.yaml

# 创建数据目录
RUN mkdir -p ./data

EXPOSE 8080

# 启动命令
CMD ["./teamspeak-backend"]
