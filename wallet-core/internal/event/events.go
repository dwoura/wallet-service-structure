package event

// WithdrawalCreatedEvent 提现创建事件
// Topic: wallet_events_withdrawal
type WithdrawalCreatedEvent struct {
	WithdrawalID uint64 `json:"withdrawal_id"`
	UserID       uint64 `json:"user_id"`
	ToAddress    string `json:"to_address"`
	Amount       string `json:"amount"` // Decimal string
	Chain        string `json:"chain"`
}
