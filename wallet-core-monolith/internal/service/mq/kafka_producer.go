package mq

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer 实现 Producer 接口
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer 创建 Kafka 生产者
// brokers: Kafka 节点地址列表 (e.g. ["localhost:9092"])
// topic: 默认主题 (如果没有指定)
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	// 配置 Writer
	// 关键配置 (面试点):
	// 1. Balancer: 决定消息发往哪个 Partition (默认 RoundRobin, 但指定 Key 后按 Key hash)
	// 2. Async: 默认是异步批量发送，提高吞吐量
	// 3. RequiredAcks: 决定可靠性级别 (None, One, All)
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},    // 按 Key 哈希，保证同一用户的消息有序
		AllowAutoTopicCreation: true,             // 开发环境允许自动创建 Topic
		RequiredAcks:           kafka.RequireAll, // 强一致性: 等待所有 ISR 副本确认
		BatchSize:              100,              // 批量发送大小
		BatchTimeout:           10 * time.Millisecond,
	}

	return &KafkaProducer{
		writer: writer,
	}
}

// Publish 发送消息到 Kafka
func (p *KafkaProducer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	// 构造消息
	msg := kafka.Message{
		// Topic: topic, // Writer 已指定 Topic，此处不需要再指定，否则报错
		Value: payload,
		Key:   []byte(key), // 使用传入的 Key 保证分区有序
	}

	// 发送 (底层是异步批量的，但在 Writer 层面是阻塞等待 Ack)
	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("[Kafka] Publish Error: %v", err)
		return fmt.Errorf("kafka write error: %w", err)
	}

	return nil
}

// Close 关闭连接
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
