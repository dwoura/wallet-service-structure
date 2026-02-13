package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, val, ttl).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string, target interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss")
	}
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), target)
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
