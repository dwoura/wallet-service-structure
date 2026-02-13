package routes

import (
	"wallet-core/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(rg *gin.RouterGroup) {
	adminGroup := rg.Group("/admin")
	// 可以在这里添加 AdminAuth 中间件
	{
		adminGroup.POST("/withdrawals/:id/review", handler.Admin.ReviewWithdrawal)
	}
}
