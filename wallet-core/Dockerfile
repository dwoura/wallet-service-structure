# Build Stage
FROM golang:1.24-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

WORKDIR /app

# 预下载依赖 (利用 Docker 缓存)
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译 (静态链接)
RUN CGO_ENABLED=0 GOOS=linux go build -o wallet-server ./cmd/wallet-server
RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/user-service
RUN CGO_ENABLED=0 GOOS=linux go build -o wallet-service ./cmd/wallet-service
RUN CGO_ENABLED=0 GOOS=linux go build -o broadcaster-worker ./cmd/broadcaster-worker

# Run Stage
FROM alpine:latest

# 安装基础运行依赖
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/wallet-server .
COPY --from=builder /app/user-service .
COPY --from=builder /app/wallet-service .
COPY --from=builder /app/broadcaster-worker .

# 暴露端口
# 8080: Gateway/Monolith HTTP
# 50051: Monolith gRPC
# 50052: Wallet Service gRPC
# 50053: User Service gRPC
EXPOSE 8080 50051 50052 50053

# 默认运行 wallet-server (可被 K8s command 覆盖)
CMD ["./wallet-server"]
