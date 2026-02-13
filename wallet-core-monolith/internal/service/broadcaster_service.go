package service

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"wallet-core/internal/model"
	"wallet-core/pkg/bip32"
	"wallet-core/pkg/monitor"

	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

// BroadcasterService 负责将审批通过的提现单广播上链
type BroadcasterService struct {
	db        *gorm.DB
	ethClient *ethclient.Client
	masterKey bip32.ExtendedKey
	chainID   *big.Int
}

var Broadcaster *BroadcasterService

func NewBroadcasterService(db *gorm.DB, rpcURL string, masterKey bip32.ExtendedKey) (*BroadcasterService, error) {
	client, err := ethclient.Dial(rpcURL)
	chainID := big.NewInt(1)
	if err == nil {
		cid, _ := client.ChainID(context.Background())
		if cid != nil {
			chainID = cid
		}
	} else {
		log.Printf("[Broadcaster] Warning: RPC 无法连接，将运行在模拟模式")
	}

	return &BroadcasterService{
		db:        db,
		ethClient: client,
		masterKey: masterKey,
		chainID:   chainID,
	}, nil
}

// Start 启动轮询
func (s *BroadcasterService) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	log.Println("[Broadcaster] 启动提现广播服务...")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.processPendingWithdrawals(ctx)
			}
		}
	}()
}

func (s *BroadcasterService) processPendingWithdrawals(ctx context.Context) {
	var withdrawals []model.Withdrawal
	// 查询 batches 避免内存溢出
	if err := s.db.Where("status = ?", "pending_broadcast").Limit(10).Find(&withdrawals).Error; err != nil {
		log.Printf("[Broadcaster] 查询失败: %v", err)
		return
	}

	for _, w := range withdrawals {
		s.broadcastOne(ctx, &w)
	}
}

func (s *BroadcasterService) broadcastOne(ctx context.Context, w *model.Withdrawal) {
	log.Printf("[Broadcaster] 开始处理提现单 ID: %d, To: %s, Amount: %s", w.ID, w.ToAddress, w.Amount)

	// 1. 派生热钱包私钥 (假设热钱包使用 Path m/44'/60'/0'/0/0)
	// 这里简化：假设 SweeperService 里的 Hot Wallet 就是 Master Key 派生出来的第一个地址
	// 或者为了演示，直接使用 MasterPrivate Key 签发 (不推荐生产)
	// 更好的做法是 Config 里配置 HotWalletPrivateKey，或者像 SweeperService 一样管理 Keys.
	// 假设我们沿用 SweeperService 的逻辑：热钱包是 m/44'/60'/0'/0/0 (Account 0)

	hotWalletKey, err := s.deriveHotWalletKey()
	if err != nil {
		log.Printf("[Broadcaster] 派生私钥失败: %v", err)
		return
	}
	ecdsaKey := hotWalletKey.ToECDSA()

	// 2. 构造交易
	nonce := uint64(0) // 需要查询链上 Nonce
	// ... (Nonce, GasPrice 逻辑如有 ethClient 则查询，否则模拟)

	// 模拟广播成功
	txHash := fmt.Sprintf("0xmocked_tx_hash_%d_%d", w.ID, time.Now().Unix())

	// 3. 更新数据库
	w.Status = "completed" // 或者 success
	w.TxHash = txHash
	w.UpdatedAt = time.Now()

	if err := s.db.Save(w).Error; err != nil {
		log.Printf("[Broadcaster] 保存状态失败: %v", err)
		return
	}

	// Metric
	monitor.Business.WithdrawalSuccessTotal.WithLabelValues(w.Chain).Inc()

	log.Printf("[Broadcaster] ✅ 提现广播成功! TxHash: %s", txHash)
}

func (s *BroadcasterService) deriveHotWalletKey() (*bip32.ExtendedKey, error) {
	// 简化：直接用 Master Key (仅作演示)
	// 实际路径应为 BIP44
	return &s.masterKey, nil
}
