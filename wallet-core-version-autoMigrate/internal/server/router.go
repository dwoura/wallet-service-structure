package server

import (
	"wallet-core/internal/handler"
	"wallet-core/internal/handler/response"

	"wallet-core/pkg/monitor"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewHTTPRouter 初始化并返回一个 Gin Engine
func NewHTTPRouter() *gin.Engine {
	// 0. 初始化监控指标
	monitor.Init()

	// 1. 创建 Engine (使用默认中间件: Logger, Recovery)
	r := gin.Default()

	// 2. 注册通用中间件
	r.Use(monitor.PrometheusMiddleware()) // [NEW] 监控埋点

	// 3. 注册基础路由
	r.GET("/health", handler.HealthCheck)
	r.GET("/metrics", gin.WrapH(promhttp.Handler())) // [NEW] 暴露给 Prometheus
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 4. 注册 API 路由组
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			response.Success(c, gin.H{"pong": true})
		})

		// 未来在这里添加更多业务路由，例如:
		// wallet := api.Group("/wallet")
		// wallet.POST("/deposit_address", walletHandler.GetDepositAddress)
	}

	return r
}
