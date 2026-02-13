package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"wallet-core/internal/event"
	"wallet-core/internal/model"
	"wallet-core/internal/service"
	"wallet-core/internal/service/mq"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	ErrAccountNotFound = errors.New("账户不存在")
	ErrInsufficient    = errors.New("余额不足")
)

type Service struct {
	db       *gorm.DB
	addrSvc  service.AddressService // 依赖 AddressService 生成地址
	producer mq.Producer            // 依赖 MQ Producer 发送提现事件
}

func NewService(db *gorm.DB, addrSvc service.AddressService, producer mq.Producer) *Service {
	return &Service{
		db:       db,
		addrSvc:  addrSvc,
		producer: producer,
	}
}

// CreateAddress 为用户生成充值地址
func (s *Service) CreateAddress(ctx context.Context, userID int64, currency string) (string, error) {
	// 调用 AddressService 生成地址 (这里复用现有逻辑，未来可以将 AddressService 也拆分)
	// 注意: 这里的 userID 转为 uint64 适配旧接口
	addr, _, err := s.addrSvc.GetDepositAddress(uint64(userID), currency)
	if err != nil {
		return "", err
	}
	return addr, nil
}

// GetBalance 获取用户余额
func (s *Service) GetBalance(ctx context.Context, userID int64, currency string) (map[string]string, error) {
	var accounts []model.Account
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)

	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	if err := query.Find(&accounts).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, acc := range accounts {
		// 返回总金额 (可用 + 冻结) 或者仅可用，视业务需求。这里返回总额。
		total := acc.Balance.Add(acc.LockedBalance)
		result[acc.Currency] = total.String()
	}
	return result, nil
}

// CreateWithdrawal 创建提现申请
func (s *Service) CreateWithdrawal(ctx context.Context, userID int64, toAddr, amountStr, currency string) (int64, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return 0, errors.New("金额格式错误")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return 0, errors.New("提现金额必须大于0")
	}

	// 开启事务 (检查余额 -> 扣除余额 -> 创建提现记录)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var account model.Account
		// 悲观锁: SELECT ... FOR UPDATE
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ? AND currency = ?", userID, currency).
			First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAccountNotFound
			}
			return err
		}

		if account.Balance.LessThan(amount) {
			return ErrInsufficient
		}

		// 扣钱 (Balance -> LockedBalance)
		// 简化: 直接扣除 Balance，生成 Withdrawal 记录 (状态 pending)
		// 严谨做法: Balance -= amount, LockedBalance += amount
		account.Balance = account.Balance.Sub(amount)
		account.LockedBalance = account.LockedBalance.Add(amount)

		if err := tx.Save(&account).Error; err != nil {
			return err
		}

		// TODO: 创建 withdraw_records 记录 (这里仅演示核心逻辑，假设 withdrawal 表存在)
		// withdrawal := model.Withdrawal{...}
		// tx.Create(&withdrawal)

		return nil
	})

	if err != nil {
		return 0, err
	}

	// 返回模拟的 ID (实际中应该是 withdrawal.ID)
	withdrawalID := time.Now().UnixNano()

	// 发送提现创建事件 (Async)
	// Topic: wallet_events_withdrawal
	go func() {
		payload, _ := json.Marshal(event.WithdrawalCreatedEvent{
			WithdrawalID: uint64(withdrawalID), // Mock ID
			UserID:       uint64(userID),
			ToAddress:    toAddr,
			Amount:       amountStr,
			Chain:        currency, // Assuming currency maps to chain for now
		})
		// 使用 UserID 作为 Partition Key 保证顺序
		_ = s.producer.Publish(context.Background(), "wallet_events_withdrawal", strconv.FormatInt(userID, 10), payload)
	}()

	return withdrawalID, nil
}
