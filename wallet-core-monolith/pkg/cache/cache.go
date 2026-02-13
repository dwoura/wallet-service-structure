package cache

import (
	"context"
	"time"
)

// Cache 定义通用缓存接口
type Cache interface {
	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Get 获取缓存，并将结果 Unmarshal 到 target 中
	Get(ctx context.Context, key string, target interface{}) error
	// Delete 删除缓存
	Delete(ctx context.Context, key string) error
}
