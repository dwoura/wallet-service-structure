package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Collection 资金归集记录
type Collection struct {
	ID        uint `gorm:"primaryKey"`
	DepositID uint `gorm:"uniqueIndex;not null"` // 关联的充值记录 ID (一对一，防止重复归集)

	// 交易信息
	TxHash      string          `gorm:"type:varchar(66);uniqueIndex;not null"`
	FromAddress string          `gorm:"type:varchar(42);not null"`
	ToAddress   string          `gorm:"type:varchar(42);not null"`
	Amount      decimal.Decimal `gorm:"type:decimal(30,0);not null"` // 实际归集金额 (余额 - Gas)
	GasFee      decimal.Decimal `gorm:"type:decimal(30,0);not null"` // 消耗的 Gas 费

	// 状态
	Status string `gorm:"type:varchar(20);not null;default:'pending'"` // pending, confirmed, failed

	CreatedAt time.Time
	UpdatedAt time.Time
}
