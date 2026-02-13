package safe_random

import (
	"encoding/hex"
	"math/big"
	"testing"
)

func TestGenerateRandomBytes(t *testing.T) {
	n := 32
	b, err := GenerateRandomBytes(n)
	if err != nil {
		t.Fatalf("GenerateRandomBytes 失败: %v", err)
	}
	if len(b) != n {
		t.Errorf("GenerateRandomBytes 返回了 %d 字节, 期望 %d", len(b), n)
	}

	// 简单的随机性检查（极不可能全为零）
	allZero := true
	for _, v := range b {
		if v != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Error("GenerateRandomBytes 返回了全零数据，可能未正确生成随机数")
	}
}

func TestGenerateRandomHexString(t *testing.T) {
	n := 16
	s, err := GenerateRandomHexString(n)
	if err != nil {
		t.Fatalf("GenerateRandomHexString 失败: %v", err)
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("解码 Hex 字符串失败: %v", err)
	}

	if len(decoded) != n {
		t.Errorf("GenerateRandomHexString 底层字节长度 = %d, 期望 %d", len(decoded), n)
	}
}

func TestGenerateRandomInt(t *testing.T) {
	max := big.NewInt(100)
	for i := 0; i < 100; i++ {
		n, err := GenerateRandomInt(max)
		if err != nil {
			t.Fatalf("GenerateRandomInt 失败: %v", err)
		}
		if n.Cmp(big.NewInt(0)) < 0 || n.Cmp(max) >= 0 {
			t.Errorf("GenerateRandomInt 返回值 %v 超出范围 [0, %v)", n, max)
		}
	}
}
