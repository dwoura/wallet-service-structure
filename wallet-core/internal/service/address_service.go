package service

import (
	"context"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"wallet-core/internal/model"
	"wallet-core/pkg/address"
	"wallet-core/pkg/bip32"
	"wallet-core/pkg/cache"
)

// SQLAddressService 是 AddressService 的实现
type SQLAddressService struct {
	db          *gorm.DB
	redis       *redis.Client
	masterKey   bip32.ExtendedKey // 系统级的主公钥 (xpub)，用于派生子地址
	btcGen      *address.BTCGenerator
	ethGen      *address.ETHGenerator
	networkType string // "mainnet" or "testnet"
	cache       cache.Cache
}

// NewSQLAddressService 构造函数
func NewSQLAddressService(db *gorm.DB, rdb *redis.Client, masterKey bip32.ExtendedKey, network *chaincfg.Params, c cache.Cache) (*SQLAddressService, error) {
	// 确保 masterKey 是公钥 (为了安全，服务层不需要私钥)
	if masterKey.IsPrivate() {
		return nil, fmt.Errorf("AddressService 应该只持有扩展公钥 (xpub)，不应持有私钥")
	}

	return &SQLAddressService{
		db:          db,
		redis:       rdb,
		masterKey:   masterKey,
		btcGen:      address.NewBTCGenerator(network),
		ethGen:      address.NewETHGenerator(),
		networkType: network.Name,
		cache:       c,
	}, nil
}

// GetDepositAddress 获取充值地址
// 逻辑:
// 1. 查询 DB 是否已存在该链的地址 (同一个用户每条链目前只分配一个充值地址)
// 2. 如果不存在，从 Redis 获取下一个 path_index
// 3. 派生子公钥 -> 生成地址
// 4. 保存到 DB
// 5. 返回地址
func (s *SQLAddressService) GetDepositAddress(uid uint64, chain string) (string, int, error) {
	// 1. 查库
	var existingAddr model.Address
	err := s.db.Where("user_id = ? AND chain = ?", uid, chain).First(&existingAddr).Error
	if err == nil {
		// 找到了
		return existingAddr.Address, existingAddr.HDPathIndex, nil
	}
	if err != gorm.ErrRecordNotFound {
		// 数据库错误
		return "", 0, fmt.Errorf("数据库查询错误: %w", err)
	}

	// 2. 生成新地址
	// 使用 Redis INCR 原子递增作为 HD Path Index
	// Key: wallet:hd_index:{chain}
	redisKey := fmt.Sprintf("wallet:hd_index:%s", chain)
	index, err := s.redis.Incr(context.Background(), redisKey).Result()
	if err != nil {
		return "", 0, fmt.Errorf("Redis INCR 失败: %w", err)
	}
	hdPathIndex := int(index)

	// 3. 派生密钥
	// 假设我们使用 BIP-44 路径的简化版: m/purpose'/coin_type'/0'/0/index
	// 由于我们这里只持有 xpub，我们无法进行 hardened derivation (如 m/44')
	// 所以我们的 masterKey 必须已经是 m/44'/coin_type'/0' 这一级的 account key
	// 或者我们简化: 直接从 root xpub 派生 m/0/index (非标准但简单)

	// 为了演示标准做法，假设传入的 masterKey 已经是 Account Extended Public Key
	// 例如: m/44'/0'/0' (BTC Account 0)
	// 那么我们只需要派生 m/0/index (External Chain / Index)

	// 这里为了简化演示，我们假设 masterKey 是 Root Xpub，我们只做非硬化派生
	// 派生路径: m/0/index (0=External Chain)
	// 注意: 实际生产中应严格管理 xpub 的层级

	// 派生路径: 0 (external chain)
	chainKey, err := s.masterKey.Derive(0)
	if err != nil {
		return "", 0, fmt.Errorf("密钥派生失败 (chain): %w", err)
	}
	// 派生路径: index
	childKey, err := chainKey.Derive(uint32(hdPathIndex))
	if err != nil {
		return "", 0, fmt.Errorf("密钥派生失败 (index): %w", err)
	}

	// 转换公钥
	ecPubKey, err := childKey.ECPubKey()
	if err != nil {
		return "", 0, fmt.Errorf("获取 EC 公钥失败: %w", err)
	}

	var addressStr string
	if chain == "BTC" {
		addressStr, err = s.btcGen.PubKeyToAddress(ecPubKey.SerializeCompressed())
	} else if chain == "ETH" {
		addressStr, err = s.ethGen.PubKeyToAddress(ecPubKey.SerializeUncompressed())
	} else {
		return "", 0, fmt.Errorf("不支持的链: %s", chain)
	}

	if err != nil {
		return "", 0, fmt.Errorf("地址生成失败: %w", err)
	}

	// 4. 保存到 DB
	newAddr := model.Address{
		UserID:      uid,
		Chain:       chain,
		Address:     addressStr,
		HDPathIndex: hdPathIndex,
		CreatedAt:   time.Now(),
	}

	// 处理并发: 唯一索引冲突 (有可能两个请求同时拿到同一个 Redis index? 不可能，Redis 是原子的)
	// 但有可能同一个用户并发请求，导致这里重复插入
	// 使用 FirstOrCreate 或者 事务
	err = s.db.Create(&newAddr).Error
	if err != nil {
		// 如果是唯一索引冲突，说明已经有一条了，再查一次返回即可
		// 这里简化处理
		return "", 0, fmt.Errorf("保存地址到数据库失败: %w", err)
	}

	return addressStr, hdPathIndex, nil
}

// GetSupportedCurrencies 获取支持的币种列表 (演示缓存)
func (s *SQLAddressService) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	cacheKey := "wallet:supported_currencies"
	var currencies []string

	// 1. 查缓存 (L1 -> L2)
	// 注意: 我们在 Cache 接口定义了 Get(ctx, key, target)
	if err := s.cache.Get(ctx, cacheKey, &currencies); err == nil {
		return currencies, nil
	}

	// 2. 缓存未名中 (模拟查 DB 或 Config)
	// 真实场景可能是: s.db.Find(&currencies)
	currencies = []string{"BTC", "ETH", "TRON", "USDT"}

	// 3. 写缓存 (TTL 1小时)
	_ = s.cache.Set(ctx, cacheKey, currencies, time.Hour)

	return currencies, nil
}
