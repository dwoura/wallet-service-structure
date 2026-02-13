package main

import (
	"context"
	"fmt"
	"time"

	"wallet-core/api/pb"
	handler_grpc "wallet-core/internal/handler/grpc"
	"wallet-core/internal/model"
	"wallet-core/internal/server"
	"wallet-core/internal/service"
	"wallet-core/internal/service/mq"
	"wallet-core/internal/service/observer"

	"wallet-core/pkg/bip32"
	"wallet-core/pkg/bip39"
	"wallet-core/pkg/config"
	"wallet-core/pkg/database"
	"wallet-core/pkg/logger"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	_ "wallet-core/docs/swagger"
)

// @title Wallet Core API
// @version 1.0
// @description BlockChain Wallet Server API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 0. 初始化 Config
	config.Init()

	// 1. 初始化 Logger
	logger.Init(config.Global.App.Env)
	defer logger.Sync()

	// 2. 构造 DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Global.DB.Host,
		config.Global.DB.User,
		config.Global.DB.Password,
		config.Global.DB.Name,
		config.Global.DB.Port,
	)

	// 模拟从配置加载的助记词
	mnemonic := config.Global.Wallet.Mnemonic

	// 2. 连接数据库
	db, err := database.ConnectPostgres(dsn)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}

	// 3. 连接 Redis
	rdb, err := database.ConnectRedis(config.Global.Redis.Addr, config.Global.Redis.Password, config.Global.Redis.DB)
	if err != nil {
		logger.Fatal("Redis 连接失败", zap.Error(err))
	}

	// 4. 执行数据库迁移 (Auto Migrate) - [DEPRECATED]
	if config.Global.App.Env == "development" {
		logger.Info("开发环境: 尝试自动迁移 Schema (GORM AutoMigrate)...")
		err = db.AutoMigrate(model.AllModels()...)
		if err != nil {
			logger.Fatal("数据库自动迁移失败", zap.Error(err))
		}
		logger.Info("数据库自动迁移完成 (Dev Mode)")
	} else {
		logger.Info("生产环境: 跳过 AutoMigrate，请使用 migrate 工具管理 Schema")
	}

	// 5. 初始化核心钱包模块
	// 5.1 生成 Seed
	mnemonicService := bip39.NewMnemonicService()
	seed := mnemonicService.MnemonicToSeed(mnemonic, "")

	// 5.2 生成 Master Key
	wallet, err := bip32.NewMasterKeyFromSeed(seed, &chaincfg.MainNetParams)
	if err != nil {
		logger.Fatal("生成 Master Key 失败", zap.Error(err))
	}
	masterKey := wallet.MasterKey()
	logger.Info("Master Key (XPrv) 加载成功 (内存中)")

	// 5.3 转换为主公钥 (XPub)
	masterPub, err := masterKey.Neuter()
	if err != nil {
		logger.Fatal("转换 XPub 失败", zap.Error(err))
	}
	logger.Info("Master Key (XPub)", zap.String("xpub", masterPub.String()))

	// 6. 初始化充值地址服务
	addressService, err := service.NewSQLAddressService(db, rdb, masterPub, &chaincfg.MainNetParams)
	if err != nil {
		logger.Fatal("初始化 AddressService 失败", zap.Error(err))
	}

	// 7. 测试业务逻辑
	testUser(db, addressService)

	// 8. 初始化消息队列
	mqType := config.Global.Redis.MQType
	var producer mq.Producer
	var consumer mq.Consumer

	if mqType == "kafka" {
		logger.Info("使用 Kafka 作为消息队列...")
		kafkaBrokers := config.Global.Kafka.Brokers
		producer = mq.NewKafkaProducer(kafkaBrokers, "wallet_events_deposit")
		consumer = mq.NewKafkaConsumer(kafkaBrokers, "wallet_sweeper_group")
	} else {
		logger.Info("使用 Redis Streams 作为消息队列...")
		producer = mq.NewRedisProducer(rdb)
		consumer = mq.NewRedisConsumer(rdb, "wallet_sweeper", "sweeper-0")
	}

	// 9. 启动消息中继服务
	relayService := service.NewRelayService(db, producer)
	go relayService.Start(context.Background())

	// 10. 启动区块扫描器
	ethObserver := observer.NewEthObserver(db, producer, 3000, 5)
	go func() {
		if err := ethObserver.Start(context.Background()); err != nil {
			logger.Error("Observer 启动失败", zap.Error(err))
		}
	}()

	// 11. 启动资金归集服务
	hotWallet := config.Global.Wallet.HotWallet
	rpcURL := config.Global.Wallet.RpcUrl
	sweeper, err := service.NewSweeperService(db, consumer, rpcURL, masterKey, hotWallet, rdb)
	if err != nil {
		logger.Error("Sweeper 初始化失败", zap.Error(err))
	} else {
		go func() {
			if err := sweeper.Start(context.Background()); err != nil {
				logger.Error("Sweeper 运行出错", zap.Error(err))
			}
		}()
	}

	// ==========================================
	// Server Startup (Refactored)
	// ==========================================

	// 12. HTTP Router
	// 路由注册已被合并进 server 包
	r := server.NewHTTPRouter()

	// 13. gRPC Server
	grpcServer := grpc.NewServer()
	pb.RegisterAddressServiceServer(grpcServer, handler_grpc.NewAddressHandler(addressService))

	// 14. 启动应用
	cfg := server.Config{
		HttpPort: config.Global.App.HttpPort,
		GrpcPort: config.Global.App.GrpcPort,
	}

	// 初始化 Server App
	app, err := server.New(cfg, r, grpcServer)
	if err != nil {
		logger.Fatal("应用启动失败", zap.Error(err))
	}

	// 运行 (阻塞)
	app.Run()

	// 15. 退出后资源清理
	logger.Info("正在关闭数据库连接...")
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}
	rdb.Close()
	logger.Info("系统已退出")
}

func testUser(db *gorm.DB, addressService service.AddressService) {
	var count int64
	db.Model(&model.User{}).Count(&count)

	var user model.User
	if count == 0 {
		logger.Info("插入测试用户...")
		user = model.User{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: "hashed",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		db.Create(&user)
		db.Create(&model.Account{UserID: user.ID, Currency: "BTC", Balance: decimal.Zero})
		db.Create(&model.Account{UserID: user.ID, Currency: "ETH", Balance: decimal.Zero})
	} else {
		db.First(&user)
	}

	// 为该用户生成 BTC 地址
	btcAddr, idx, err := addressService.GetDepositAddress(user.ID, "BTC")
	if err != nil {
		logger.Error("生成 BTC 地址失败", zap.Error(err))
	} else {
		logger.Info("BTC 充值地址",
			zap.String("username", user.Username),
			zap.Uint64("uid", uint64(user.ID)),
			zap.String("address", btcAddr),
			zap.Int("index", idx))
	}

	// 为该用户生成 ETH 地址
	ethAddr, idx, err := addressService.GetDepositAddress(user.ID, "ETH")
	if err != nil {
		logger.Error("生成 ETH 地址失败", zap.Error(err))
	} else {
		logger.Info("ETH 充值地址",
			zap.String("username", user.Username),
			zap.Uint64("uid", uint64(user.ID)),
			zap.String("address", ethAddr),
			zap.Int("index", idx))
	}
}
