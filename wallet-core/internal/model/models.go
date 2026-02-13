package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"type:varchar(255);not null;unique" json:"username"`
	Email        string         `gorm:"type:varchar(255);not null;unique" json:"email"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"` // 不返回密码
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Accounts []Account `gorm:"foreignKey:UserID" json:"accounts,omitempty"`
}

// Account 资产账户表
// 核心设计: 引入 Version 字段实现乐观锁
type Account struct {
	ID            uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64          `gorm:"not null;index" json:"user_id"`
	Currency      string          `gorm:"type:varchar(10);not null;uniqueIndex:idx_user_currency" json:"currency"` // BTC, ETH
	Balance       decimal.Decimal `gorm:"type:decimal(32,18);not null;default:0" json:"balance"`                   // 总余额
	LockedBalance decimal.Decimal `gorm:"type:decimal(32,18);not null;default:0" json:"locked_balance"`            // 冻结余额
	Version       uint64          `gorm:"not null;default:0" json:"version"`                                       // 乐观锁版本号
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Account) TableName() string {
	return "accounts"
}

// OutboxMessage 本地消息表 (Transactional Outbox)
type OutboxMessage struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Topic     string         `gorm:"type:varchar(255);not null" json:"topic"`
	Payload   []byte         `gorm:"type:text;not null" json:"payload"`
	Status    string         `gorm:"type:varchar(50);not null;default:'PENDING';index" json:"status"` // PENDING, SENT, FAILED
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OutboxMessage) TableName() string {
	return "outbox_messages"
}
