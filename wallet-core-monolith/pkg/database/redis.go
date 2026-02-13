package database

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis 连接到 Redis
// addr: "localhost:6379"
// password: ""
func ConnectRedis(addr string, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("无法连接到 Redis: %w", err)
	}

	log.Println("Redis 连接成功")
	return rdb, nil
}
