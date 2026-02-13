package crypto_util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/sha3"
	"lukechampine.com/blake3"
)

// CalculateMD5 计算输入的 MD5 哈希值。
// 警告：MD5 不安全，不应用于安全相关的用途。
func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// CalculateSHA256 计算输入的 SHA256 哈希值。
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// CalculateKeccak256 计算输入的 Keccak256 哈希值。
// 这是以太坊使用的哈希算法。
func CalculateKeccak256(data []byte) string {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// CalculateBlake3 计算输入的 Blake3 哈希值。
// Blake3 是一种现代、高性能的加密哈希函数。
func CalculateBlake3(data []byte) string {
	hash := blake3.Sum256(data)
	return hex.EncodeToString(hash[:])
}
