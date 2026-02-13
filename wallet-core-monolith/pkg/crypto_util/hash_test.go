package crypto_util

import (
	"testing"
)

func TestHashes(t *testing.T) {
	input := []byte("hello world")

	// MD5
	md5Hash := CalculateMD5(input)
	if len(md5Hash) != 32 { // 16 bytes * 2 hex chars
		t.Errorf("MD5 哈希长度不匹配: 得到 %d, 期望 32", len(md5Hash))
	}
	t.Logf("MD5: %s", md5Hash)

	// SHA256
	sha256Hash := CalculateSHA256(input)
	if len(sha256Hash) != 64 { // 32 bytes * 2 hex chars
		t.Errorf("SHA256 哈希长度不匹配: 得到 %d, 期望 64", len(sha256Hash))
	}
	t.Logf("SHA256: %s", sha256Hash)

	// Keccak256
	keccakHash := CalculateKeccak256(input)
	if len(keccakHash) != 64 {
		t.Errorf("Keccak256 哈希长度不匹配: 得到 %d, 期望 64", len(keccakHash))
	}
	t.Logf("Keccak256: %s", keccakHash)

	// Blake3
	blake3Hash := CalculateBlake3(input)
	if len(blake3Hash) != 64 {
		t.Errorf("Blake3 哈希长度不匹配: 得到 %d, 期望 64", len(blake3Hash))
	}
	t.Logf("Blake3: %s", blake3Hash)
}
