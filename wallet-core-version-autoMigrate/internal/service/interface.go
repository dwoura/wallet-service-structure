package service

type AddressService interface {
	// GetDepositAddress 为用户获取或生成指定链的充值地址
	// uid: 用户ID
	// chain: 链名称 (BTC, ETH)
	// 返回: (地址字符串, path_index, error)
	GetDepositAddress(uid uint64, chain string) (string, int, error)
}
