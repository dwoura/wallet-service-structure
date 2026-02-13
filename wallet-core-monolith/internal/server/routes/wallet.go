package routes

import (
	"wallet-core/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterWalletRoutes(rg *gin.RouterGroup) {
	walletGroup := rg.Group("/wallet")
	// Auth middleware here
	{
		walletGroup.POST("/withdraw", handler.Withdraw.CreateWithdrawal)
	}
}
