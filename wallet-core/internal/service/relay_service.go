package service

import (
	"context"
	"log"
	"time"

	"wallet-core/internal/model"
	"wallet-core/internal/service/mq"

	"gorm.io/gorm"
)

// RelayService 负责将本地消息表的消息搬运到 MQ
type RelayService struct {
	db       *gorm.DB
	producer mq.Producer
	interval time.Duration
}

func NewRelayService(db *gorm.DB, producer mq.Producer) *RelayService {
	return &RelayService{
		db:       db,
		producer: producer,
		interval: 500 * time.Millisecond, // 500ms 轮询一次
	}
}

func (s *RelayService) Start(ctx context.Context) {
	log.Println("[Relay] 启动消息中继服务...")
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[Relay] 停止服务")
			return
		case <-ticker.C:
			s.processPendingMessages(ctx)
		}
	}
}

func (s *RelayService) processPendingMessages(ctx context.Context) {
	// 1. 获取一批 Pending 消息
	var messages []model.OutboxMessage
	// 每次取 50 条，避免内存爆炸
	if err := s.db.Where("status = ?", "PENDING").Limit(50).Find(&messages).Error; err != nil {
		log.Printf("[Relay] 查询消息失败: %v", err)
		return
	}

	if len(messages) == 0 {
		return
	}

	log.Printf("[Relay] 发现 %d 条待发送消息", len(messages))

	for _, msg := range messages {
		// 2. 发送 MQ
		// Key: 这里简化处理，可以把 Key 也存在表里，或者从 Payload 解析
		key := ""
		if err := s.producer.Publish(ctx, msg.Topic, key, msg.Payload); err != nil {
			log.Printf("[Relay] 发送消息 ID=%d 失败: %v", msg.ID, err)
			// 可以增加重试次数计数
			continue
		}

		// 3. 更新状态为 SENT
		// 只有发送成功了才更新状态 => At-least-once (至少一次投递)
		// 如果这里更新失败，下次还会发，Consumer 需做好幂等
		if err := s.db.Model(&msg).Update("status", "SENT").Error; err != nil {
			log.Printf("[Relay] 更新状态 ID=%d 失败: %v", msg.ID, err)
		} else {
			log.Printf("[Relay] 消息 ID=%d 已投递", msg.ID)
		}
	}
}
