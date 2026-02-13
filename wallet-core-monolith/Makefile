.PHONY: all build test clean run up down db-shell

APP_NAME=wallet-cli

all: build

# 编译 CLI 工具
build:
	go build -o $(APP_NAME) ./cmd/wallet-cli

# 运行所有单元测试
test:
	go test -v ./...

# 清理编译产物
clean:
	rm -f $(APP_NAME)
	rm -f wallet.db # 如果有 SQLite 文件

# 运行 CLI (示例)
run-help:
	./$(APP_NAME) help

# 启动基础设施 (Docker Compose)
up:
	docker-compose up -d

# 停止基础设施
down:
	docker-compose down

# 进入数据库命令行 (Port 5433)
db-shell:
	docker exec -it wallet-db psql -U wallet_user -d wallet_db

# 进入 Redis 命令行 (Port 6380)
redis-shell:
	docker exec -it wallet-redis redis-cli
