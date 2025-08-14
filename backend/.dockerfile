# backend/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o teamspeak-monitor .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/teamspeak-monitor .
COPY --from=builder /app/config.yaml .

EXPOSE 8080
CMD ["./teamspeak-monitor"]
