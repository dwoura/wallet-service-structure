package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestHealthCheck 这是一个集成测试示例
// 它假设 Wallet Server 已经在运行 (例如通过 Docker Compose)
// 运行命令: go test -v ./tests/integration/...
func TestHealthCheck(t *testing.T) {
	// 1. 设置目标 URL (通常从环境变量读取)
	baseURL := "http://localhost:8080/api/v1"

	// 2. 发起请求
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/ping")

	// 3. 断言结果
	if err != nil {
		t.Skip("Skipping integration test: server not running? " + err.Error())
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
