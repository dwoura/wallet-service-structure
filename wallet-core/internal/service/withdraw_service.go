package service

import (
	"context"

	"wallet-core/internal/model"
	"wallet-core/pkg/database"

	"gorm.io/gorm"
)

type WithdrawService struct{}

var Withdraw = &WithdrawService{}

// CreateWithdrawal 创建提现申请
func (s *WithdrawService) CreateWithdrawal(ctx context.Context, userID uint64, req *model.Withdrawal) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 检查余额 (这里暂时略过，假设足够)
		// ...

		// 2. 设置初始状态 (关键点: pending_review)
		req.Status = "pending_review"
		req.RequiredApprovals = 2 // 从 Config 读取，这里硬编码演示
		req.CurrentApprovals = 0

		// 3. 创建记录
		if err := tx.Create(req).Error; err != nil {
			return err
		}

		return nil
	})
}
