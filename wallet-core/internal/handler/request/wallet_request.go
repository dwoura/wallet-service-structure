package request

// CreateDepositAddressRequest 生成充值地址请求
type CreateDepositAddressRequest struct {
	UserID   uint64 `json:"user_id" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=BTC ETH"`
}
