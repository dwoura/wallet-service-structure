package mq

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisProducer 实现 Producer 接口
type RedisProducer struct {
	client *redis.Client
}

// NewRedisProducer 创建 Redis 生产者
func NewRedisProducer(client *redis.Client) *RedisProducer {
	return &RedisProducer{
		client: client,
	}
}

// Publish 发送消息到 Redis Stream
func (p *RedisProducer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	// 使用 Redis Streams 的 XADD 命令
	// Stream Name = topic (e.g., "wallet:events:deposit")

	err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		Values: map[string]interface{}{
			"payload": payload,
			// 可以添加更多元数据，如 timestamp
		},
	}).Err()

	if err != nil {
		log.Printf("[MQ] Publish Error: %v", err)
		return fmt.Errorf("redis xadd error: %w", err)
	}

	return nil
}

// RedisConsumer 实现 Consumer 接口
type RedisConsumer struct {
	client *redis.Client
	group  string
	name   string
}

// NewRedisConsumer 创建 Redis 消费者
func NewRedisConsumer(client *redis.Client, group, name string) *RedisConsumer {
	return &RedisConsumer{
		client: client,
		group:  group,
		name:   name,
	}
}

// Subscribe 订阅 Redis Stream
func (c *RedisConsumer) Subscribe(ctx context.Context, topic string, handler func(msg *Message) error) error {
	// 1. 创建 Consumer Group (如果不存在)
	// XGROUP CREATE <stream> <group> $ MKSTREAM
	err := c.client.XGroupCreateMkStream(ctx, topic, c.group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("创建消费者组失败: %w", err)
	}

	log.Printf("[Redis MQ] 开始监听主题: %s (Group: %s)", topic, c.group)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// 2. 阻塞读取消息
			// XREADGROUP GROUP <group> <consumer> BLOCK 2000 COUNT 1 STREAMS <topic> >
			streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.group,
				Consumer: c.name,
				Streams:  []string{topic, ">"},
				Count:    1,
				Block:    2 * time.Second,
			}).Result()

			if err == redis.Nil {
				continue // 超时无消息
			}
			if err != nil {
				log.Printf("[Redis MQ] 读取消息错误: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// 3. 处理消息
			for _, stream := range streams {
				for _, xMessage := range stream.Messages {
					// 解析 Payload
					val, ok := xMessage.Values["payload"].(string)
					if !ok {
						log.Printf("[Redis MQ] 消息格式错误: payload 缺失")
						c.ack(ctx, topic, xMessage.ID)
						continue
					}

					msg := &Message{
						ID:      xMessage.ID,
						Topic:   topic,
						Payload: []byte(val),
					}

					// 调用回调
					if err := handler(msg); err != nil {
						log.Printf("[Redis MQ] 消息处理失败: %v", err)
						// 可以在这里实现死信队列逻辑
					} else {
						// 处理成功，ACK
						c.ack(ctx, topic, xMessage.ID)
					}
				}
			}
		}
	}
}

func (c *RedisConsumer) ack(ctx context.Context, topic, id string) {
	c.client.XAck(ctx, topic, c.group, id)
}

func (c *RedisConsumer) Close() error {
	return c.client.Close()
}
