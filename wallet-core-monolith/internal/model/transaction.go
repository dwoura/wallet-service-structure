package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Address 充值地址表
type Address struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64    `gorm:"not null;index" json:"user_id"`
	Chain       string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_chain_address;uniqueIndex:idx_chain_path" json:"chain"` // bitcoin, ethereum
	Address     string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_chain_address" json:"address"`
	HDPathIndex int       `gorm:"not null;uniqueIndex:idx_chain_path" json:"hd_path_index"` // BIP-44 address_index
	CreatedAt   time.Time `json:"created_at"`
}

// Deposit 充值记录表
type Deposit struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64          `gorm:"not null;index" json:"user_id"`
	BlockAppID  uint64          `gorm:"not null;index" json:"block_app_id"` // 关联 Address.ID
	TxHash      string          `gorm:"type:varchar(255);not null;uniqueIndex:idx_tx_app" json:"tx_hash"`
	Amount      decimal.Decimal `gorm:"type:decimal(32,18);not null" json:"amount"`
	BlockHeight uint64          `gorm:"not null" json:"block_height"`
	Status      string          `gorm:"type:varchar(20);not null" json:"status"` // pending, confirmed
	CreatedAt   time.Time       `json:"created_at"`
	ConfirmedAt *time.Time      `json:"confirmed_at,omitempty"`
}

// Withdrawal 提现记录表
type Withdrawal struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            uint64          `gorm:"not null;index" json:"user_id"`
	ToAddress         string          `gorm:"type:varchar(255);not null" json:"to_address"`
	Amount            decimal.Decimal `gorm:"type:decimal(32,18);not null" json:"amount"`
	Chain             string          `gorm:"type:varchar(20);not null" json:"chain"`
	TxHash            string          `gorm:"type:varchar(255)" json:"tx_hash"`                                 // 提现发出后的 Hash
	Status            string          `gorm:"type:varchar(32);not null;default:'pending_review'" json:"status"` // pending_review, pending_broadcast, completed, failed
	RequiredApprovals int             `gorm:"not null;default:2" json:"required_approvals"`
	CurrentApprovals  int             `gorm:"not null;default:0" json:"current_approvals"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// WithdrawalReview 提现审核记录表
type WithdrawalReview struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	WithdrawalID uint64    `gorm:"not null;uniqueIndex:idx_withdrawal_admin" json:"withdrawal_id"`
	AdminID      uint64    `gorm:"not null;uniqueIndex:idx_withdrawal_admin" json:"admin_id"`
	Status       string    `gorm:"type:varchar(16);not null" json:"status"` // approve, reject
	Remark       string    `gorm:"type:text" json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
}

func (WithdrawalReview) TableName() string {
	return "withdrawal_reviews"
}

func (Address) TableName() string {
	return "addresses"
}

func (Deposit) TableName() string {
	return "deposits"
}

func (Withdrawal) TableName() string {
	return "withdrawals"
}
