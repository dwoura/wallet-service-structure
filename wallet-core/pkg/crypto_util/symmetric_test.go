package crypto_util

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestAESGCM(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 字节用于 AES-256
	plaintext := []byte("这是一条用于 AES-GCM 测试的秘密消息")

	ciphertext, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptAESGCM 失败: %v", err)
	}

	decrypted, err := DecryptAESGCM(key, ciphertext)
	if err != nil {
		t.Fatalf("DecryptAESGCM 失败: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("解密后的消息与明文不匹配。\n得到: %s\n期望: %s", decrypted, plaintext)
	}
}

func TestDES(t *testing.T) {
	key := []byte("12345678") // DES 使用 8 字节
	plaintext := []byte("Secret!!")

	ciphertext, err := EncryptDES(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptDES 失败: %v", err)
	}

	decrypted, err := DecryptDES(key, ciphertext)
	if err != nil {
		t.Fatalf("DecryptDES 失败: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("解密后的消息与明文不匹配。\n得到: %s\n期望: %s", decrypted, plaintext)
	}
}

func TestAESGCM_InvalidKey(t *testing.T) {
	key := []byte("shortkey")
	plaintext := []byte("test")
	_, err := EncryptAESGCM(key, plaintext)
	if err == nil {
		t.Error("期望因密钥长度无效而报错，但未收到错误")
	}
}

func TestDES_InvalidKey(t *testing.T) {
	key := []byte("wrong")
	plaintext := []byte("test")
	_, err := EncryptDES(key, plaintext)
	if err == nil {
		t.Error("期望因 DES 密钥长度无效而报错，但未收到错误")
	}
}

func TestVectors(t *testing.T) {
	// Simple sanity check with a known hex string to ensure no panic
	// Note: Since IV/Nonce is random, we can't check against a fixed ciphertext strictly without mocking rand.
	// But we can check that we produce valid hex.
	key := []byte("0123456789abcdef") // 16 bytes
	res, err := EncryptAESGCM(key, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("AES Ciphertext (Hex): %s", hex.EncodeToString(res))
}
