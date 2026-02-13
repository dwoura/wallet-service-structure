package request

import "github.com/shopspring/decimal"

type CreateWithdrawalRequest struct {
	ToAddress string          `json:"to_address" binding:"required"`
	Amount    decimal.Decimal `json:"amount" binding:"required"`
	Chain     string          `json:"chain" binding:"required"`
}
