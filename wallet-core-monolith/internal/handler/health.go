package handler

import (
	"wallet-core/internal/handler/response"

	"github.com/gin-gonic/gin"
)

// HealthCheck godoc
// @Summary Check system health
// @Description Get the current health status of the server
// @Tags system
// @Accept  json
// @Produce  json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "UP",
		"version": "1.0.0",
		"service": "wallet-server",
	})
}
