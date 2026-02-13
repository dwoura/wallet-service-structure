package gateway

import (
	"context"
	"net/http"
	"strconv"
	"time"

	userv1 "wallet-core/api/gen/user/v1"
	walletv1 "wallet-core/api/gen/wallet/v1"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all HTTP routes for the gateway
func RegisterRoutes(r *gin.Engine, userClient userv1.UserServiceClient, walletClient walletv1.WalletServiceClient) {
	api := r.Group("/v1")

	// User Routes
	userHandler := &UserHandler{client: userClient}
	api.POST("/user/register", userHandler.Register)
	api.POST("/user/login", userHandler.Login)
	api.GET("/user/profile", userHandler.GetProfile)

	// Wallet Routes
	walletHandler := &WalletHandler{client: walletClient}
	api.POST("/wallet/address", walletHandler.CreateAddress)
	api.GET("/wallet/balance", walletHandler.GetBalance)
	api.POST("/wallet/withdraw", walletHandler.CreateWithdrawal)
}

type UserHandler struct {
	client userv1.UserServiceClient
}

func (h *UserHandler) Register(c *gin.Context) {
	var req userv1.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.Register(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req userv1.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.Login(ctx, &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id must be a valid number"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetUserInfo(ctx, &userv1.GetUserInfoRequest{UserId: userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

type WalletHandler struct {
	client walletv1.WalletServiceClient
}

func (h *WalletHandler) CreateAddress(c *gin.Context) {
	var req walletv1.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CreateAddress(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	currency := c.Query("currency")

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id must be a valid number"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetBalance(ctx, &walletv1.GetBalanceRequest{
		UserId:   userID,
		Currency: currency,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WalletHandler) CreateWithdrawal(c *gin.Context) {
	var req walletv1.CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CreateWithdrawal(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
