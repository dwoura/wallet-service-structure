package observer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"wallet-core/internal/model"
	"wallet-core/internal/service/mq"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// EthObserver 实现 ChainObserver 接口
// 核心设计:
// 1. Fetcher (生产者): 单线程，负责按顺序获取区块
// 2. Worker Pool (消费者): 多线程，负责并行处理区块内的交易
type EthObserver struct {
	db            *gorm.DB
	currentHeight uint64
	wg            sync.WaitGroup

	// 配置
	startHeight uint64
	workerCount int

	// 通道 (Channel) 作为队列
	// Fetcher -> blocksChan -> Workers
	blocksChan chan *Block

	// 消息队列生产者
	producer mq.Producer
}

// NewEthObserver 创建一个新的 ETH 扫描器
func NewEthObserver(db *gorm.DB, producer mq.Producer, startHeight uint64, workerCount int) *EthObserver {
	return &EthObserver{
		db:          db,
		startHeight: startHeight,
		workerCount: workerCount,
		// 创建带缓冲的 Channel，模拟队列
		blocksChan: make(chan *Block, workerCount*2),
		// 初始化 MQ Producer
		producer: producer,
	}
}

// Start 启动扫描器
func (o *EthObserver) Start(ctx context.Context) error {
	log.Printf("启动 ETH 扫描器，起始高度: %d, Worker 数量: %d", o.startHeight, o.workerCount)
	o.currentHeight = o.startHeight

	// 1. 启动 Workers (消费者)
	for i := 0; i < o.workerCount; i++ {
		o.wg.Add(1)
		go o.worker(ctx, i)
	}

	// 2. 启动 Fetcher (生产者)
	o.wg.Add(1)
	go o.fetcher(ctx)

	return nil
}

// Stop 停止扫描器 (空实现，因为 Start 中的 ctx 控制了退出)
func (o *EthObserver) Stop() error {
	return nil
}

func (o *EthObserver) GetCurrentHeight() uint64 {
	return o.currentHeight
}

// fetcher (生产者): 模拟从节点获取区块
func (o *EthObserver) fetcher(ctx context.Context) {
	defer o.wg.Done()
	// 当 fetcher 退出时，关闭 channel，通知 workers 全部下班
	defer close(o.blocksChan)

	ticker := time.NewTicker(1 * time.Second) // 模拟 1秒一个块
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Fetcher: 收到退出信号，停止获取区块")
			return
		case <-ticker.C:
			// 模拟获取一个新的区块
			block := o.mockFetchBlock(o.currentHeight)

			// 将任务发送给 Workers
			// 注意: 如果 Worker 处理不过来，这里会阻塞，从而实现"背压" (Backpressure)
			select {
			case o.blocksChan <- block:
				log.Printf("Fetcher: 推送区块 #%d 到处理队列", block.Height)
				o.currentHeight++
			case <-ctx.Done():
				return
			}
		}
	}
}

// worker (消费者): 处理区块中的交易
func (o *EthObserver) worker(ctx context.Context, id int) {
	defer o.wg.Done()
	log.Printf("Worker-%d: 上线待命", id)

	for block := range o.blocksChan {
		// 模拟处理耗时
		// time.Sleep(500 * time.Millisecond)

		// 检查区块中的每一笔交易
		for _, tx := range block.Transactions {
			// 在这里实现真正的业务逻辑:
			// 1. 检查 To 地址是否在我们的 addresses 表中
			// 2. 如果在，插入 deposits 表 (加锁或原子操作)

			o.processTransaction(tx)
		}

		log.Printf("Worker-%d: 完成区块 #%d 的处理", id, block.Height)
	}

	log.Printf("Worker-%d: 队列已关闭，下班", id)
}

// processTransaction 业务逻辑核心
func (o *EthObserver) processTransaction(tx Transaction) {
	// 1. 检查 To 地址是否是我们的用户地址
	// 使用带缓存的查询或布隆过滤器会更好，这里先用 DB 直查
	var count int64
	// 只查 ETH 链的地址
	err := o.db.Model(&model.Address{}).
		Where("address = ? AND chain = ?", tx.To, "ETH").
		Count(&count).Error

	if err != nil {
		log.Printf("  [Error] 查询地址失败: %v", err)
		return
	}

	if count > 0 {
		// 2. 命中！这是充值交易
		log.Printf("  [$$$] 发现充值交易! Tx: %s, To: %s, Amount: %s", tx.Hash, tx.To, tx.Value)

		// 3. 开启事务 (Transactional Outbox 核心)
		err = o.db.Transaction(func(dbTx *gorm.DB) error {
			// A. 写入 Deposit 表
			deposit := model.Deposit{
				UserID:      1, // Hack
				BlockAppID:  0, // Hack
				TxHash:      tx.Hash,
				Amount:      decimal.RequireFromString(tx.Value),
				BlockHeight: o.currentHeight,
				Status:      "confirmed",
				CreatedAt:   time.Now(),
			}

			// 完善逻辑
			var addr model.Address
			dbTx.Where("address = ?", tx.To).First(&addr)
			deposit.UserID = addr.UserID
			deposit.BlockAppID = addr.ID

			if err := dbTx.Create(&deposit).Error; err != nil {
				return err // 回滚
			}

			// B. 写入 Outbox 消息表 (在同一个事务中!)
			payloadMap := map[string]interface{}{
				"user_id": deposit.UserID,
				"amount":  deposit.Amount.String(),
				"tx_hash": deposit.TxHash,
				"chain":   "ETH",
			}

			// Topic: wallet_events_deposit
			if err := model.CreateOutboxMessage(dbTx, "wallet_events_deposit", payloadMap); err != nil {
				return err // 回滚
			}

			return nil // 提交事务
		})

		if err != nil {
			log.Printf("  [Error] 充值事务处理失败: %v", err)
		} else {
			log.Printf("  [Success] 充值事务提交成功 (Deposit + Outbox)")
		}
	}
}

// mockFetchBlock 模拟生成测试区块
func (o *EthObserver) mockFetchBlock(height uint64) *Block {
	// 模拟构造一些随机交易
	// 偶尔生成一笔发给我们的测试地址: 0x40ceeEdE9fA9ee09e594aFFb63CFc4994aF5B14e
	targetAddr := "0xRandom"
	if height%5 == 0 { // 每5个块生成一笔真实充值
		targetAddr = "0x40ceeEdE9fA9ee09e594aFFb63CFc4994aF5B14e"
	}

	txs := []Transaction{
		{Hash: fmt.Sprintf("0xhash_%d_1", height), To: targetAddr, Value: "0.5"},
		{Hash: fmt.Sprintf("0xhash_%d_2", height), To: "0xSomeoneElse", Value: "100"},
	}
	return &Block{
		Height:       height,
		Hash:         fmt.Sprintf("0xblock_%d", height),
		Transactions: txs,
	}
}
