package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"wallet-core/pkg/logger"
	"wallet-core/pkg/utils/lock"
)

type CronService struct {
	cron  *cron.Cron
	redis *redis.Client
}

func NewCronService(rdb *redis.Client) *CronService {
	// 使用秒级调度 (By default cron/v3 is minute level, use Ensure standard or use WithSeconds)
	// 这里使用标准配置 (分级)
	c := cron.New()
	return &CronService{
		cron:  c,
		redis: rdb,
	}
}

func (s *CronService) Start() {
	// 注册任务
	_, _ = s.cron.AddFunc("@every 1m", s.SyncExchangeRates) // 每分钟同步汇率

	s.cron.Start()
	logger.Info("Cron Service started")
}

func (s *CronService) Stop() {
	s.cron.Stop()
	logger.Info("Cron Service stopped")
}

// SyncExchangeRates 模拟同步汇率任务
func (s *CronService) SyncExchangeRates() {
	ctx := context.Background()
	lockKey := "cron:lock:sync_rates"

	// 1. 获取分布式锁 (TTL 10s)
	// 防止多实例同时执行
	locker := lock.NewRedisLock(s.redis)
	locked, err := locker.Acquire(ctx, lockKey, 10*time.Second)
	if err != nil || !locked {
		// 获取锁失败，说明有其他节点在运行，跳过
		logger.Debug("SyncExchangeRates: 获取锁失败或已有实例在运行")
		return
	}
	defer locker.Release(ctx, lockKey)

	// 2. 执行任务逻辑
	logger.Info("开始同步汇率...", zap.String("time", time.Now().String()))
	time.Sleep(2 * time.Second) // 模拟耗时
	logger.Info("汇率同步完成")
}
