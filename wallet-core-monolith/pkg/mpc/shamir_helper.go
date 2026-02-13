package mpc

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/shamir"
)

// Split 将私钥切分为 N 个部分，至少需要 M 个才能恢复
// secretHex: 原始私钥 (Hex String)
// parts: 切分总数 (N)
// threshold: 恢复阈值 (M)
// return: N 个 Share (Hex String)
func Split(secretHex string, parts, threshold int) ([]string, error) {
	// 1. 去掉 0x 前缀并 Decode
	secretHex = strings.TrimPrefix(secretHex, "0x")
	secretBytes, err := hex.DecodeString(secretHex)
	if err != nil {
		return nil, fmt.Errorf("invalid secret hex: %v", err)
	}

	// 2. 调用 Shamir Split
	sharesBytes, err := shamir.Split(secretBytes, parts, threshold)
	if err != nil {
		return nil, err
	}

	// 3. 将每个 Share 编码为 Hex
	// Share 的第一个字节通常是 X 坐标，后面是 Y 值
	var shares []string
	for _, share := range sharesBytes {
		shares = append(shares, hex.EncodeToString(share))
	}

	return shares, nil
}

// Recover 从 M 个 Shares 中恢复私钥
// sharesHex: M 个 Share (Hex String)
// return: 原始私钥 (Hex String)
func Recover(sharesHex []string) (string, error) {
	// 1. Decode Shares
	var sharesBytes [][]byte
	for _, s := range sharesHex {
		s = strings.TrimPrefix(s, "0x")
		b, err := hex.DecodeString(s)
		if err != nil {
			return "", fmt.Errorf("invalid share hex: %v", err)
		}
		sharesBytes = append(sharesBytes, b)
	}

	// 2. 调用 Shamir Combine
	secretBytes, err := shamir.Combine(sharesBytes)
	if err != nil {
		return "", err // 比如 shares 数量不够，或者不匹配
	}

	// 3. Encode
	return hex.EncodeToString(secretBytes), nil
}
