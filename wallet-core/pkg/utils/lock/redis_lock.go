package lock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedLock 定义分布式锁接口
type DistributedLock interface {
	// Acquire 尝试获取锁
	// key: 锁的唯一标识
	// ttl: 锁的过期时间
	// 返回: (是否成功, error)
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Release 释放锁
	Release(ctx context.Context, key string) error
}

// RedisLock 基于 Redis SETNX 的实现
type RedisLock struct {
	client *redis.Client
}

func NewRedisLock(client *redis.Client) *RedisLock {
	return &RedisLock{client: client}
}

func (l *RedisLock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	// SET key value NX EX ttl
	// value 可以是随机字符串或机器ID，用于释放时校验归属 (这里简化暂不校验)
	success, err := l.client.SetNX(ctx, "lock:"+key, "1", ttl).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

func (l *RedisLock) Release(ctx context.Context, key string) error {
	// 直接删除 Key
	// 生产环境严谨做法: 需要 Lua 脚本检查 Value 是否属于自己再删除
	return l.client.Del(ctx, "lock:"+key).Err()
}
