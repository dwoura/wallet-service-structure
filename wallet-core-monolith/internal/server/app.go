package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wallet-core/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Config struct {
	HttpPort string
	GrpcPort string
}

type App struct {
	httpServer   *http.Server
	grpcServer   *grpc.Server
	grpcListener net.Listener
}

func New(cfg Config, httpHandler *gin.Engine, grpcServer *grpc.Server) (*App, error) {
	// HTTP Server
	httpSrv := &http.Server{
		Addr:    ":" + cfg.HttpPort,
		Handler: httpHandler,
	}

	// gRPC Listener
	lis, err := net.Listen("tcp", ":"+cfg.GrpcPort)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on grpc port %s: %w", cfg.GrpcPort, err)
	}

	return &App{
		httpServer:   httpSrv,
		grpcServer:   grpcServer,
		grpcListener: lis,
	}, nil
}

// Run 启动服务并阻塞，直到收到关闭信号
func (a *App) Run() {
	// 1. Start HTTP
	go func() {
		logger.Info("Starting HTTP Server", zap.String("addr", a.httpServer.Addr))
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP Server failure", zap.Error(err))
		}
	}()

	// 2. Start gRPC
	go func() {
		logger.Info("Starting gRPC Server", zap.String("addr", a.grpcListener.Addr().String()))
		if err := a.grpcServer.Serve(a.grpcListener); err != nil {
			logger.Fatal("gRPC Server failure", zap.Error(err))
		}
	}()

	// 3. Signal Handling (Blocking)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("⚠️  Shutting down server...")

	// 4. Graceful Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP Server forced to shutdown", zap.Error(err))
	}

	a.grpcServer.GracefulStop()
	logger.Info("Server exited properly")
}
