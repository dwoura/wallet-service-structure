package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type MemoryCache struct {
	c *gocache.Cache
}

func NewMemoryCache(defaultExpiration, cleanupInterval time.Duration) *MemoryCache {
	return &MemoryCache{
		c: gocache.New(defaultExpiration, cleanupInterval),
	}
}

func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 为了保持行为一致（避免引用修改），这里也建议存副本，
	// 或者简单点只存对象。为了配合 MultiLevel，我们尽量让它表现得像 "存了值"。
	m.c.Set(key, value, ttl)
	return nil
}

func (m *MemoryCache) Get(ctx context.Context, key string, target interface{}) error {
	val, found := m.c.Get(key)
	if !found {
		return fmt.Errorf("cache miss")
	}

	// 这里的难点是：val 是 interface{}，target 是 *T
	// 简单的做法是：如果 val 和 target 类型匹配，直接赋值（需要反射）
	// 或者：为了通用性和一致性，内存缓存也存 []byte (JSON)，但这浪费了内存缓存的性能优势。
	//
	// 折中方案：使用 json.Marshal/Unmarshal 进行深拷贝和类型转换
	// 虽然有序列化开销，但保证了与 Redis 行为一致，且不用处理复杂的反射赋值。

	bytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.c.Delete(key)
	return nil
}
