package mq

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumer 实现 Consumer 接口
type KafkaConsumer struct {
	brokers []string
	groupID string
	reader  *kafka.Reader
}

// NewKafkaConsumer 创建 Kafka 消费者
func NewKafkaConsumer(brokers []string, groupID string) *KafkaConsumer {
	return &KafkaConsumer{
		brokers: brokers,
		groupID: groupID,
	}
}

// Subscribe 订阅 Kafka 主题
func (c *KafkaConsumer) Subscribe(ctx context.Context, topic string, handler func(msg *Message) error) error {
	// 配置 Reader
	// 核心配置 (面试点):
	// 1. GroupID: 消费组 ID，保证同组内只有一个消费者能消费到同一分区的消息 (负载均衡)
	// 2. StartOffset: 新组从哪里开始消费? FirstOffset (最早) / LastOffset (最新)
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		GroupID:  c.groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		// CommitInterval: 1 * time.Second, // 自动提交间隔 (如果禁用自动提交，需要手动 CommitMessages)
		StartOffset: kafka.LastOffset,
	})

	log.Printf("[Kafka MQ] 开始监听主题: %s (Group: %s)", topic, c.groupID)

	// 启动消费循环
	go c.consumeLoop(ctx, topic, handler)

	return nil
}

func (c *KafkaConsumer) consumeLoop(ctx context.Context, topic string, handler func(msg *Message) error) {
	defer c.reader.Close()

	for {
		// 1. 读取消息 (阻塞直到有消息)
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // 上下文取消，退出
			}
			log.Printf("[Kafka MQ] 读取消息错误: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// 2. 构造通用消息
		msg := &Message{
			ID:      string(m.Key), // Kafka Key 作为 ID (或者使用 Offset 组合)
			Topic:   topic,
			Key:     string(m.Key),
			Payload: m.Value,
		}

		// 3. 调用业务处理函数
		if err := handler(msg); err != nil {
			log.Printf("[Kafka MQ] 业务处理失败: %v", err)
			// TODO: 实现重试逻辑或死信队列
			// 注意: Kafka 不像 RabbitMQ 那样支持单条消息 Nack 自动重回队列。
			// 通常做法是: 提交 Offset (认为已消费), 但将失败消息写入另一个 "retry_topic" 或 DB 表。
			continue
		}

		// 4. 手动提交 Offset (确认消费成功)
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("[Kafka MQ] 提交 Offset 失败: %v", err)
		}
	}
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
