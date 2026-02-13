package safe_random

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

// GenerateRandomBytes 生成指定长度的安全随机字节切片。
// 如果系统的安全随机数生成器失败，将返回错误。
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// 注意：只有读取了 len(b) 个字节，err 才为 nil。
	if err != nil {
		return nil, fmt.Errorf("生成随机字节失败: %w", err)
	}
	return b, nil
}

// GenerateRandomHexString 生成指定长度（字节数）的 URL 安全、类似 Base64 的安全随机字符串。
// 注意：实际字符串长度是 Hex 编码后的，因此长度是请求字节数的两倍。
func GenerateRandomHexString(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateRandomInt 生成一个 [0, max) 范围内的均匀随机值。
// 如果 max <= 0 则会 panic。
func GenerateRandomInt(max *big.Int) (*big.Int, error) {
	if max.Sign() <= 0 {
		return nil, fmt.Errorf("最大值必须为正数")
	}
	return rand.Int(rand.Reader, max)
}

// Reader 是一个全局共享的加密安全随机数生成器实例。
// 默认为 crypto/rand.Reader。
var Reader io.Reader = rand.Reader
