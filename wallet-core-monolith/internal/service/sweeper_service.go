package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"wallet-core/internal/model"
	"wallet-core/internal/service/mq"
	"wallet-core/pkg/bip32"
	"wallet-core/pkg/monitor"
	"wallet-core/pkg/utils/lock"
)

// SweeperService 负责资金归集
type SweeperService struct {
	db        *gorm.DB
	consumer  mq.Consumer
	ethClient *ethclient.Client
	masterKey bip32.ExtendedKey // Root XPrv
	chainID   *big.Int
	distLock  lock.DistributedLock // 分布式锁

	// 固定的热钱包地址 (接收归集资金)
	hotWalletAddr common.Address
}

// DepositEvent 对应 MQ 中的 Payload
type DepositEvent struct {
	UserID uint   `json:"user_id"`
	Amount string `json:"amount"`
	TxHash string `json:"tx_hash"`
	Chain  string `json:"chain"`
}

func NewSweeperService(db *gorm.DB, consumer mq.Consumer, rpcURL string, masterKey bip32.ExtendedKey, hotWallet string, redisClient *redis.Client) (*SweeperService, error) {
	if !masterKey.IsPrivate() {
		return nil, fmt.Errorf("SweeperService 需要私钥")
	}

	// 尝试连接 RPC
	// 为了演示，如果连接失败，我们允许 client 为 nil，并在发送时只打印日志
	client, err := ethclient.Dial(rpcURL)
	chainID := big.NewInt(1) // Default Mainnet

	if err != nil {
		log.Printf("[Sweeper] Warning: 无法连接 ETH RPC (%s): %v. 将运行在【模拟模式】", rpcURL, err)
	} else {
		cid, err := client.ChainID(context.Background())
		if err == nil {
			chainID = cid
			log.Printf("[Sweeper] 已连接 ETH 节点, ChainID: %s", chainID.String())
		}
	}

	return &SweeperService{
		db:            db,
		consumer:      consumer,
		ethClient:     client,
		masterKey:     masterKey,
		chainID:       chainID,
		hotWalletAddr: common.HexToAddress(hotWallet),
		distLock:      lock.NewRedisLock(redisClient), // 初始化锁
	}, nil
}

func (s *SweeperService) Start(ctx context.Context) error {
	log.Println("[Sweeper] 启动资金归集服务...")
	return s.consumer.Subscribe(ctx, "wallet_events_deposit", s.handleDeposit)
}

func (s *SweeperService) handleDeposit(msg *mq.Message) error {
	// 1. 解析消息
	var event DepositEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		log.Printf("[Sweeper] Error: 解析消息失败: %v", err)
		return nil // 格式错误，不再重试
	}

	if event.Chain != "ETH" {
		return nil // 暂时只处理 ETH
	}

	log.Printf("[Sweeper] 收到充值事件: User=%d, Amount=%s, Tx=%s", event.UserID, event.Amount, event.TxHash)

	// [Metric] 记录充值金额
	if amountVal, err := decimal.NewFromString(event.Amount); err == nil {
		amountFloat, _ := amountVal.Float64()
		monitor.Business.DepositAmountTotal.WithLabelValues(event.Chain).Add(amountFloat)
	}

	// 2. 分布式锁检查 (New!)
	// 锁Key: sweeper:deposit:<tx_hash>
	lockKey := fmt.Sprintf("sweeper:deposit:%s", event.TxHash)
	locked, err := s.distLock.Acquire(context.Background(), lockKey, 10*time.Minute)
	if err != nil {
		log.Printf("[Sweeper] 获取锁系统错误: %v", err)
		return err // 重试
	}
	if !locked {
		log.Printf("[Sweeper] ⚠️ 获取锁失败 (正在被其他节点处理), 跳过: %s", event.TxHash)
		return nil
	}
	// 确保处理完释放锁 (虽然有 TTL 兜底)
	defer s.distLock.Release(context.Background(), lockKey)

	// 3. 检查数据库，防止重复归集 (双重保障)
	var count int64
	s.db.Model(&model.Collection{}).Where("tx_hash = ?", event.TxHash).Count(&count)
	if count > 0 {
		log.Printf("[Sweeper] 该充值已归集过，跳过")
		return nil
	}

	// 4. 核心逻辑: 归集所有 ETH 到热钱包
	return s.sweepETH(context.Background(), &event)
}

func (s *SweeperService) sweepETH(ctx context.Context, event *DepositEvent) error {
	// [Metric] 记录归集耗时
	timer := prometheus.NewTimer(monitor.Business.SweeperJobDuration.WithLabelValues("ETH"))
	defer timer.ObserveDuration()

	// A. 获取该用户的充值地址的 Path Index
	var addr model.Address
	if err := s.db.Where("user_id = ? AND chain = ?", event.UserID, "ETH").First(&addr).Error; err != nil {
		return fmt.Errorf("找不到用户地址: %v", err)
	}

	log.Printf("[Sweeper] 准备从地址 %s 归集资金 (Path: m/0/%d)", addr.Address, addr.HDPathIndex)

	// B. 派生私钥 (Key Derivation) !! 核心安全 !!
	// 路径: Master -> 0 (External) -> Index
	chainKey, err := s.masterKey.Derive(0)
	if err != nil {
		return err
	}
	childKey, err := chainKey.Derive(uint32(addr.HDPathIndex))
	if err != nil {
		return err
	}
	privKey, err := childKey.ECPrivKey()
	if err != nil {
		return err
	}
	// 转换为 ECDSA 私钥
	ecdsaPrivateKey := privKey.ToECDSA()

	// C. 查询余额 & 估算 Gas
	// 如果是模拟模式，我们假设余额就是充值金额
	// 如果是真实模式，查链
	balanceWei := big.NewInt(0)
	nonce := uint64(0)
	gasPrice := big.NewInt(20000000000) // 20 Gwei default

	if s.ethClient != nil {
		// 真实查询
		fromAddr := common.HexToAddress(addr.Address)
		bal, err := s.ethClient.BalanceAt(ctx, fromAddr, nil)
		if err != nil {
			log.Printf("[Sweeper] 查询余额失败: %v", err)
			return err // 重试
		}
		balanceWei = bal

		n, err := s.ethClient.PendingNonceAt(ctx, fromAddr)
		if err != nil {
			return err
		}
		nonce = n

		gp, err := s.ethClient.SuggestGasPrice(ctx)
		if err == nil {
			gasPrice = gp
		}
	} else {
		// 模拟: 余额 = 充值金额
		amountDecimal, _ := decimal.NewFromString(event.Amount)
		balanceWei = amountDecimal.Mul(decimal.New(1, 18)).BigInt()
	}

	// D. 计算归集金额
	// Amount = Balance - (GasLimit * GasPrice)
	gasLimit := uint64(21000) // 标准转账
	gasFee := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), gasPrice)

	if balanceWei.Cmp(gasFee) <= 0 {
		log.Printf("[Sweeper] 余额不足以支付 Gas，跳过归集 (Balance: %s, Fee: %s)", balanceWei, gasFee)
		return nil // 余额不足，认为是完成（或等待更多充值）
	}

	sweepAmount := new(big.Int).Sub(balanceWei, gasFee)

	// E. 构造并签名交易
	tx := types.NewTransaction(nonce, s.hotWalletAddr, sweepAmount, gasLimit, gasPrice, nil)

	// EIP-155 签名
	signer := types.NewEIP155Signer(s.chainID)
	signedTx, err := types.SignTx(tx, signer, ecdsaPrivateKey)
	if err != nil {
		return fmt.Errorf("签名失败: %v", err)
	}

	log.Printf("[Sweeper] ✍️ 交易签名完成! Hash: %s, Amount: %s Wei", signedTx.Hash().Hex(), sweepAmount)

	// F. 广播交易
	if s.ethClient != nil {
		if err := s.ethClient.SendTransaction(ctx, signedTx); err != nil {
			log.Printf("[Sweeper] 广播失败: %v", err)
			return err
		}
		log.Printf("[Sweeper] 🚀 交易已广播!")
	} else {
		log.Printf("[Sweeper] (模拟模式) 假装广播了交易: %s", signedTx.Hash().Hex())
	}

	// G. 保存记录
	collection := model.Collection{
		DepositID:   0, // 暂时不关联，或者查出来关联
		TxHash:      signedTx.Hash().Hex(),
		FromAddress: addr.Address,
		ToAddress:   s.hotWalletAddr.Hex(),
		Amount:      decimal.NewFromBigInt(sweepAmount, 0),
		GasFee:      decimal.NewFromBigInt(gasFee, 0),
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	s.db.Create(&collection)

	return nil
}
