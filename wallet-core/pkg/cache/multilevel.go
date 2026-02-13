package cache

import (
	"context"
	"fmt"
	"time"
)

// MultiLevelCache 实现多级缓存 (L1: Memory, L2: Redis)
type MultiLevelCache struct {
	local  Cache
	remote Cache
}

func NewMultiLevelCache(local, remote Cache) *MultiLevelCache {
	return &MultiLevelCache{
		local:  local,
		remote: remote,
	}
}

func (m *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 同时写入 L1 和 L2
	// 注意：L1 的 TTL 可以比 L2 短，这里简单起见设为一样，或者 L1 设为 L2 的一半
	if err := m.local.Set(ctx, key, value, ttl/2); err != nil {
		// log error
	}
	return m.remote.Set(ctx, key, value, ttl)
}

func (m *MultiLevelCache) Get(ctx context.Context, key string, target interface{}) error {
	// 1. 查 L1
	if err := m.local.Get(ctx, key, target); err == nil {
		return nil // L1 Hit
	}

	// 2. 查 L2
	if err := m.remote.Get(ctx, key, target); err == nil {
		// L2 Hit -> 回写 L1
		// 注意: 这里回写不需要太长 TTL，防止 L1 脏数据太久
		_ = m.local.Set(ctx, key, target, time.Minute)
		return nil
	}

	return fmt.Errorf("cache miss")
}

func (m *MultiLevelCache) Delete(ctx context.Context, key string) error {
	_ = m.local.Delete(ctx, key)
	return m.remote.Delete(ctx, key)
}
