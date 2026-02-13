package service

import (
	"context"
	"errors"

	"wallet-core/internal/model"
	"wallet-core/pkg/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminService struct{}

var Admin = &AdminService{}

// ReviewWithdrawal 审核提现
func (s *AdminService) ReviewWithdrawal(ctx context.Context, txID string, adminID uint64, action string, remark string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 悲观锁读取提现单
		var w model.Withdrawal
		// 使用 tx_id 查询还是 id 查询？
		// 之前设计用 tx_id，但 transaction.go 中 ID 是 primary key，TxHash 是上链后的 hash。
		// API 路径参数 :id 应该是数据库 ID。
		// 但为了兼容之前可能用 logical ID 的想法，这里假设 path param 传的是 withdraw ID (uint64)。
		// 但函数签名 txID string，我需要确认 Handler 传进来什么。
		// 假设 Handler 传进来的是 ID (string 转换后的)。
		// 为了严谨，建议 Handler 解析为 uint64 传进来。
		// 这里先假设 txID 是数据库 ID 的字符串形式，或者是业务 ID。
		// 看着 model 定义，ID 是 uint64，TxHash 是提现后的 hash。
		// 所以 API 应该传 ID。

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&w, "id = ?", txID).Error; err != nil {
			return err
		}

		// 2. 状态检查
		if w.Status != "pending_review" {
			return errors.New("withdrawal is not in pending_review state")
		}

		// 3. 检查是否已审批
		var count int64
		tx.Model(&model.WithdrawalReview{}).
			Where("withdrawal_id = ? AND admin_id = ?", w.ID, adminID).
			Count(&count)
		if count > 0 {
			return errors.New("admin has already reviewed this withdrawal")
		}

		// 4. 执行审批逻辑
		if action == "approve" {
			w.CurrentApprovals++
			// 阈值判断
			if w.CurrentApprovals >= w.RequiredApprovals {
				w.Status = "pending_broadcast"
			}
		} else if action == "reject" {
			w.Status = "rejected"
		}

		// 5. 插入审核记录
		review := model.WithdrawalReview{
			WithdrawalID: w.ID,
			AdminID:      adminID,
			Status:       action,
			Remark:       remark,
		}
		if err := tx.Create(&review).Error; err != nil {
			return err
		}

		// 6. 保存提现单状态
		return tx.Save(&w).Error
	})
}
