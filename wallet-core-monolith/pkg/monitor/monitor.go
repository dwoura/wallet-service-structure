package monitor

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTPRequestsTotal 记录 HTTP 请求总量
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration 记录 HTTP 请求耗时 (Histogram)
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency distributions.",
			Buckets: []float64{0.1, 0.3, 0.5, 1.0, 2.0, 5.0}, // 关键耗时桶
		},
		[]string{"method", "path"},
	)
)

// Init 初始化并注册监控指标
func Init() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	// 初始化业务指标
	InitBusinessMetrics()
}

// PrometheusMiddleware returns a gin middleware for monitoring
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // 使用路由模板 /api/v1/user/:id 而不是具体路径

		// 处理请求
		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// 记录指标
		if path != "" { // 忽略 404 等未匹配路由
			HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
			HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
		}
	}
}
