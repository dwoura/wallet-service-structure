package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

// CreateOutboxMessage 在同一个事务中创建业务数据和 Outbox 消息
func CreateOutboxMessage(tx *gorm.DB, topic string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := OutboxMessage{
		Topic:   topic,
		Payload: payloadBytes,
		Status:  "PENDING",
	}

	return tx.Create(&msg).Error
}
