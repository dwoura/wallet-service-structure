package worker

import (
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"wallet-core/internal/worker/tasks"
	"wallet-core/pkg/logger"
)

// Server 封装 Asynq Server (Worker)
type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

// NewServer 初始化 Worker Server
func NewServer(addr string, password string, db int, concurrency int) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     addr,
			Password: password,
			DB:       db,
		},
		asynq.Config{
			// 并发数：同时处理多少个任务
			Concurrency: concurrency,
			// 队列优先级
			Queues: map[string]int{
				"critical": 6, // 高优先级
				"default":  3, // 默认
				"low":      1, // 低优先级
			},
			// 错误日志处理
			Logger: logger.NewAsynqLogger(),
		},
	)

	mux := asynq.NewServeMux()

	// 注册任务处理器
	mux.HandleFunc(tasks.TypeEmailDelivery, tasks.HandleEmailDeliveryTask)

	return &Server{
		server: srv,
		mux:    mux,
	}
}

// Run 启动 Worker (阻塞)
func (s *Server) Run() error {
	logger.Info("Worker Server starting...")
	return s.server.Run(s.mux)
}

// Start 非阻塞启动 (用于集成到 main.go)
func (s *Server) Start() {
	go func() {
		if err := s.server.Run(s.mux); err != nil {
			logger.Fatal("Worker Server failed", zap.Error(err))
		}
	}()
}

// Stop 停止 Worker
func (s *Server) Stop() {
	s.server.Stop()
	s.server.Shutdown()
}
