package routes

import (
	"github.com/gin-gonic/gin"

	"wallet-core/internal/handler"
)

// RegisterUserRoutes 注册用户模块路由
func RegisterUserRoutes(rg *gin.RouterGroup) {
	// 用户相关路由
	// POST /api/v1/register
	rg.POST("/register", handler.Register)
}
