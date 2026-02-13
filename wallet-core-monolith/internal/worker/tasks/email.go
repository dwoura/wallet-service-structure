package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"wallet-core/pkg/logger"
)

// 任务类型常量
const (
	TypeEmailDelivery = "email:deliver"
)

// EmailDeliveryPayload 邮件任务参数
type EmailDeliveryPayload struct {
	UserID  uint64 `json:"user_id"`
	Subject string `json:"subject"`
}

// ---------------------------------------------------------------------
// 1. Producer (Client) Code
// ---------------------------------------------------------------------

// NewEmailDeliveryTask 创建邮件发送任务
func NewEmailDeliveryTask(userID uint64, subject string) (*asynq.Task, error) {
	payload, err := json.Marshal(EmailDeliveryPayload{UserID: userID, Subject: subject})
	if err != nil {
		return nil, err
	}
	// 默认 30 分钟超时，最多重试 5 次
	return asynq.NewTask(TypeEmailDelivery, payload, asynq.MaxRetry(5), asynq.Timeout(30*time.Minute)), nil
}

// ---------------------------------------------------------------------
// 2. Consumer (Server) Code
// ---------------------------------------------------------------------

// HandleEmailDeliveryTask 处理邮件发送任务
func HandleEmailDeliveryTask(ctx context.Context, t *asynq.Task) error {
	var p EmailDeliveryPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		// JSON 解析失败，重试也没用，直接跳过 (SkipRetry)
		// 任务会进入 Archived 队列，方便排查
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	logger.Info("开始处理邮件任务",
		zap.Uint64("user_id", p.UserID),
		zap.String("subject", p.Subject),
	)

	// 模拟耗时操作: 发送邮件
	time.Sleep(2 * time.Second)

	// 模拟随机失败 (用于测试重试机制)
	// if time.Now().Unix()%3 == 0 {
	// 	return fmt.Errorf("模拟网络超时")
	// }

	logger.Info("邮件发送成功", zap.Uint64("user_id", p.UserID))
	return nil
}
