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

# Run Stage
FROM alpine:latest

# 安装基础运行依赖
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/wallet-server .

# 暴露端口 (虽然 wallet-server 目前没有监听 HTTP 端口，但这是一个好习惯)
# EXPOSE 8080

# 运行
CMD ["./wallet-server"]
