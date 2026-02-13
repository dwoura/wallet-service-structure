package main

import (
	"net/http"

	"wallet-core/internal/gateway"
	"wallet-core/pkg/config"
	"wallet-core/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userv1 "wallet-core/api/gen/user/v1"
	walletv1 "wallet-core/api/gen/wallet/v1"
)

func main() {
	// 1. Init Config & Logger
	config.Init()
	logger.Init(config.Global.App.Env)
	defer logger.Sync()

	logger.Info("Starting Blockchain Gateway...", zap.String("port", config.Global.App.HttpPort))

	// 2. Connect to gRPC Services
	// In production, use Service Discovery (K8s Service Name)
	// User Service
	userConn, err := grpc.Dial("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to User Service", zap.Error(err))
	}
	defer userConn.Close()
	userClient := userv1.NewUserServiceClient(userConn)
	logger.Info("Connected to User Service at localhost:50053")

	// Wallet Service
	walletConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("Failed to connect to Wallet Service", zap.Error(err))
	}
	defer walletConn.Close()
	walletClient := walletv1.NewWalletServiceClient(walletConn)
	logger.Info("Connected to Wallet Service at localhost:50052")

	// 3. Init HTTP Server (Gin)
	r := gin.Default()

	// 4. Setup Routes
	gateway.RegisterRoutes(r, userClient, walletClient)

	// 5. Start Server
	srv := &http.Server{
		Addr:    ":" + config.Global.App.HttpPort,
		Handler: r,
	}

	logger.Info("Gateway listening on HTTP", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Gateway start failed", zap.Error(err))
	}
}
