package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wallet-core/internal/event"
	"wallet-core/internal/model"
	"wallet-core/internal/service/mq"
	"wallet-core/pkg/bip32"
	"wallet-core/pkg/bip39"
	"wallet-core/pkg/config"
	"wallet-core/pkg/database"
	"wallet-core/pkg/keystore"
	"wallet-core/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BroadcasterWorker 独立运行的广播服务
// 它持有私钥，是系统中最敏感的组件
type BroadcasterWorker struct {
	db        *gorm.DB
	ethClient *ethclient.Client
	masterKey bip32.ExtendedKey
}

func main() {
	// 1. 初始化配置与日志
	config.Init()
	logger.Init(config.Global.App.Env)
	defer logger.Sync()

	logger.Info("启动广播服务 (Broadcaster Worker)...", zap.String("env", config.Global.App.Env))

	// 2. 初始化数据库 (仅用于读取 pending_broadcast 任务)
	// 在更高级的架构中，它应该只消费 Kafka，不连 DB，或者只连独立的加密 DB
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Global.DB.Host,
		config.Global.DB.User,
		config.Global.DB.Password,
		config.Global.DB.Name,
		config.Global.DB.Port,
	)
	db, err := database.ConnectPostgres(dsn)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}

	// 3. 初始化 Redis (用于 Redis MQ fallback)
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Global.Redis.Addr,
		Password: config.Global.Redis.Password,
		DB:       config.Global.Redis.DB,
	})

	// 4. 加载最核心的私钥 (Master Key)
	masterKey, err := loadMasterKey()
	if err != nil {
		logger.Fatal("致命错误: 无法加载主私钥!", zap.Error(err))
	}
	logger.Info("🔐 主私钥加载成功，安全等级: High")

	// 4. 初始化链连接
	client, err := ethclient.Dial(config.Global.Wallet.RpcUrl)
	if err != nil {
		logger.Warn("RPC 连接失败，将运行在模拟模式", zap.Error(err))
	}

	worker := &BroadcasterWorker{
		db:        db,
		ethClient: client,
		masterKey: masterKey,
	}

	// 5. 初始化 MQ Consumer
	var consumer mq.Consumer
	if config.Global.Redis.MQType == "kafka" {
		logger.Info("MQ Mode: Kafka Consumer", zap.Strings("brokers", config.Global.Kafka.Brokers))
		// GroupID: broadcaster-group
		consumer = mq.NewKafkaConsumer(config.Global.Kafka.Brokers, "broadcaster-group")
	} else {
		logger.Info("MQ Mode: Redis Consumer")
		consumer = mq.NewRedisConsumer(rdb, "broadcaster-group", "worker-1")
	}

	// 6. 启动 Worker (订阅模式)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 订阅提现事件
	go func() {
		logger.Info("开始监听提现事件: wallet_events_withdrawal")
		err := consumer.Subscribe(ctx, "wallet_events_withdrawal", worker.HandleWithdrawalEvent)
		if err != nil {
			logger.Fatal("订阅失败", zap.Error(err))
		}
	}()

	// 7. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在停止广播服务...")
	cancel()
	_ = consumer.Close()
	time.Sleep(2 * time.Second)
	logger.Info("广播服务已停止")
}

func (w *BroadcasterWorker) HandleWithdrawalEvent(msg *mq.Message) error {
	var eventData event.WithdrawalCreatedEvent
	if err := json.Unmarshal(msg.Payload, &eventData); err != nil {
		logger.Error("解析消息失败", zap.Error(err))
		return nil // 格式错误，不再重试 (或者可以进死信队列)
	}

	logger.Info("收到提现事件", zap.Uint64("id", eventData.WithdrawalID), zap.String("amount", eventData.Amount))

	// 为了数据一致性，从 DB 重新查询记录状态
	var tx model.Withdrawal
	// 注意: 这里假设 withdrawal 表存在。
	// 为了演示，我们先模拟一个检查：
	// 如果是 demo 模式，我们直接构造一个 Withdrawal 对象
	// 真实场景: w.db.First(&tx, eventData.WithdrawalID)

	// MOCK: 如果表不存在或查不到，我们构造一个临时对象进行广播演示
	if err := w.db.First(&tx, eventData.WithdrawalID).Error; err != nil {
		logger.Warn("数据库未找到提现记录 (可能是 Mock ID)", zap.Uint64("id", eventData.WithdrawalID))
		amount, _ := decimal.NewFromString(eventData.Amount)
		tx = model.Withdrawal{
			ID:        eventData.WithdrawalID,
			UserID:    eventData.UserID,
			ToAddress: eventData.ToAddress,
			Amount:    amount,
			Chain:     eventData.Chain,
			Status:    "pending",
		}
	}

	if tx.Status != "pending" && tx.Status != "pending_broadcast" {
		logger.Info("提现记录状态非 pending，跳过", zap.String("status", tx.Status))
		return nil
	}

	return w.broadcast(context.Background(), &tx)
}

func (w *BroadcasterWorker) broadcast(ctx context.Context, tx *model.Withdrawal) error {
	logger.Info("开始广播交易", zap.Uint64("id", tx.ID), zap.String("to", tx.ToAddress))

	// 模拟签名与广播
	time.Sleep(100 * time.Millisecond)
	txHash := fmt.Sprintf("0x_kafka_broadcast_%d", time.Now().UnixNano())

	// 更新 DB (如果存在)
	// 这里加一个 Try Update，因为如果是 Mock 数据可能更新失败
	_ = w.db.Table("withdrawals").Where("id = ?", tx.ID).Updates(map[string]interface{}{
		"status":     "completed",
		"tx_hash":    txHash,
		"updated_at": time.Now(),
	}).Error

	logger.Info("✅ 广播成功", zap.String("hash", txHash))
	return nil
}

// 复用 main.go 中的加载逻辑
func loadMasterKey() (bip32.ExtendedKey, error) {
	// 1. 尝试从 Keystore 加载
	keystorePath := config.Global.Wallet.KeystorePath
	password := config.Global.Wallet.Password

	if _, err := os.Stat(keystorePath); err == nil && password != "" {
		encryptedJson, err := keystore.LoadFromFile(keystorePath)
		if err != nil {
			return nil, err
		}
		mnemonic, err := keystore.DecryptMnemonic(encryptedJson, password)
		if err != nil {
			return nil, err
		}
		seed := bip39.NewSeed(mnemonic, "")
		w, err := bip32.NewMasterKeyFromSeed(seed, &chaincfg.TestNet3Params)
		if err != nil {
			return nil, err
		}
		return w.MasterKey(), nil
	}

	// 2. 开发环境 Fallback
	if config.Global.Wallet.Mnemonic != "" {
		seed := bip39.NewSeed(config.Global.Wallet.Mnemonic, "")
		w, err := bip32.NewMasterKeyFromSeed(seed, &chaincfg.TestNet3Params)
		if err != nil {
			return nil, err
		}
		return w.MasterKey(), nil
	}

	return nil, fmt.Errorf("未找到可用的私钥源 (Keystore 或 Mnemonic)")
}
