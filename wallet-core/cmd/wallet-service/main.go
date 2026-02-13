package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	walletv1 "wallet-core/api/gen/wallet/v1"
	"wallet-core/cmd/wallet-service/server"
	"wallet-core/internal/service"
	"wallet-core/internal/service/mq"
	"wallet-core/internal/service/wallet"
	"wallet-core/pkg/bip32"
	"wallet-core/pkg/bip39"
	"wallet-core/pkg/cache"
	"wallet-core/pkg/config"
	"wallet-core/pkg/database"
	"wallet-core/pkg/keystore"
	"wallet-core/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. 初始化配置
	config.Init()

	// 2. 初始化日志
	logger.Init(config.Global.App.Env)
	defer logger.Sync()

	logger.Info("正在启动钱包服务 (Wallet Service)...", zap.String("env", config.Global.App.Env))

	// 3. 初始化数据库连接
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Global.DB.Host,
		config.Global.DB.User,
		config.Global.DB.Password,
		config.Global.DB.Name,
		config.Global.DB.Port,
	)
	db, err := database.ConnectPostgres(dsn)
	if err != nil {
		logger.Fatal("无法连接到数据库", zap.Error(err))
	}

	// 4. 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Global.Redis.Addr,
		Password: config.Global.Redis.Password,
		DB:       config.Global.Redis.DB,
	})

	// 5. 加载 Master Key (用于地址生成)
	// 注意: AddressService 只需要 xpub (扩展公钥)，但 loadMasterKey 返回的是私钥
	// 为了兼容现有代码，我们先加载私钥，NewSQLAddressService 内部会检查
	masterKey, err := loadMasterKey()
	if err != nil {
		logger.Fatal("无法加载主密钥", zap.Error(err))
	}

	// 6. 初始化 Cache
	// 使用本地内存缓存 (go-cache)
	c := cache.NewMemoryCache(5*time.Minute, 10*time.Minute)

	// 7. 初始化 Network Params
	// 假设我们在 Testnet
	netParams := &chaincfg.TestNet3Params

	// 8. 初始化 MQ Producer
	var producer mq.Producer
	if config.Global.Redis.MQType == "kafka" {
		logger.Info("MQ Mode: Kafka", zap.Strings("brokers", config.Global.Kafka.Brokers))
		// Topic 默认为 wallet_events_withdrawal，但在 Publish 时会指定具体 Topic
		producer = mq.NewKafkaProducer(config.Global.Kafka.Brokers, "wallet_events_default")
	} else {
		logger.Info("MQ Mode: Redis Stream")
		producer = mq.NewRedisProducer(rdb)
	}

	// 9. 初始化服务依赖
	addrSvc, err := service.NewSQLAddressService(db, rdb, masterKey, netParams, c)
	if err != nil {
		logger.Fatal("初始化 AddressService 失败", zap.Error(err))
	}

	svc := wallet.NewService(db, addrSvc, producer)

	// 9. 初始化 gRPC 服务器
	grpcServer := grpc.NewServer()
	walletServer := server.NewWalletGRPCServer(svc)
	walletv1.RegisterWalletServiceServer(grpcServer, walletServer)

	// 启用反射 (grpcurl 调试用)
	reflection.Register(grpcServer)

	// 10. 监听端口
	// Wallet service port: 50052
	port := ":50052"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatal("监听端口失败", zap.String("port", port), zap.Error(err))
	}

	logger.Info("钱包服务 (Wallet Service) 已启动 gRPC 监听", zap.String("port", port))

	// 11. 优雅停机
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("gRPC 服务异常退出", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭钱包服务...")
	grpcServer.GracefulStop()
	logger.Info("钱包服务已停止")
}

// 复用加载逻辑
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
