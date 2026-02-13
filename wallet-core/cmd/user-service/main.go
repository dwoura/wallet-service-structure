package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	userv1 "wallet-core/api/gen/user/v1"
	"wallet-core/cmd/user-service/server"
	"wallet-core/internal/service/user"
	"wallet-core/pkg/config"
	"wallet-core/pkg/database"
	"wallet-core/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Init Config
	config.Init()

	// 2. Init Logger
	logger.Init(config.Global.App.Env)
	defer logger.Sync()

	logger.Info("Starting User Service...", zap.String("env", config.Global.App.Env))

	// 3. Init DB
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		config.Global.DB.Host,
		config.Global.DB.User,
		config.Global.DB.Password,
		config.Global.DB.Name,
		config.Global.DB.Port,
	)
	db, err := database.ConnectPostgres(dsn)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 4. Init Service
	svc := user.NewService(db)

	// 5. Init gRPC Server
	grpcServer := grpc.NewServer()
	userServer := server.NewUserGRPCServer(svc)
	userv1.RegisterUserServiceServer(grpcServer, userServer)

	// Enable reflection for debugging (grpcurl)
	reflection.Register(grpcServer)

	// 6. Listen
	// User service port: 50053 (as per plan)
	port := ":50053"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatal("Failed to listen", zap.String("port", port), zap.Error(err))
	}

	logger.Info("User Service listening on gRPC", zap.String("port", port))

	// 7. Graceful Shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down User Service...")
	grpcServer.GracefulStop()
	logger.Info("User Service stopped")
}
