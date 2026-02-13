package service

import "context"

type AddressService interface {
	// GetDepositAddress 获取或生成充值地址
	// chain: BTC, ETH, TRON...
	GetDepositAddress(userID uint64, chain string) (string, int, error)

	// GetSupportedCurrencies 获取支持的币种列表 (Cached)
	GetSupportedCurrencies(ctx context.Context) ([]string, error)
}
