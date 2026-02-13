package mq

import "context"

// Message 代表一条通用的业务消息
type Message struct {
	ID       string            // 消息ID (例如 Redis Stream ID)
	Topic    string            // 主题 (例如 "deposit_events")
	Key      string            // 分区键 (例如 UserID), 同样用于 Kafka Partition
	Payload  []byte            // 消息体 (JSON)
	Metadata map[string]string // 元数据
}

// Producer 生产者接口
type Producer interface {
	// Publish 发送消息
	// key:用于分区排序 (Partition Key), 例如 UserID. 传空字符串则随机分区.
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

// Consumer 消费者接口
type Consumer interface {
	// Subscribe 订阅主题
	// handler: 消息处理函数，返回 error 会触发重试
	Subscribe(ctx context.Context, topic string, handler func(msg *Message) error) error

	// Close 关闭消费者
	Close() error
}
