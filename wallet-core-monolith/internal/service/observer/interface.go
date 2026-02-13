package observer

import "context"

// ChainObserver 定义了区块扫描器的通用行为
type ChainObserver interface {
	// Start 启动扫描器
	// ctx: 用于控制优雅退出
	Start(ctx context.Context) error

	// Stop 停止扫描器
	Stop() error

	// GetCurrentHeight 获取当前已处理到的区块高度
	GetCurrentHeight() uint64
}

// Block 包含了一个区块的基本信息 (简化版)
type Block struct {
	Height       uint64
	Hash         string
	Transactions []Transaction
}

// Transaction 包含了一笔交易的基本信息 (简化版)
type Transaction struct {
	Hash   string
	From   string
	To     string
	Value  string // Wei in Decimal String
	Status int    // 1 = Success, 0 = Fail
}
