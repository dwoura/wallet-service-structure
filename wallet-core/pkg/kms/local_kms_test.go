package kms

import (
	"bytes"
	"testing"
)

func TestLocalKMS_AES(t *testing.T) {
	kms := NewLocalKMS()

	// 创建 AES 密钥
	keyID, err := kms.CreateKey(KeyTypeAES)
	if err != nil {
		t.Fatalf("创建 AES 密钥失败: %v", err)
	}
	t.Logf("生成的 AES KeyID: %s", keyID)

	// 测试加密
	plaintext := []byte("这是最高机密")
	ciphertext, err := kms.Encrypt(keyID, plaintext)
	if err != nil {
		t.Fatalf("AES 加密失败: %v", err)
	}

	// 测试解密
	decrypted, err := kms.Decrypt(keyID, ciphertext)
	if err != nil {
		t.Fatalf("AES 解密失败: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("AES 解密内容不匹配")
	}
}

func TestLocalKMS_RSA(t *testing.T) {
	kms := NewLocalKMS()

	// 创建 RSA 密钥
	keyID, err := kms.CreateKey(KeyTypeRSA)
	if err != nil {
		t.Fatalf("创建 RSA 密钥失败: %v", err)
	}
	t.Logf("生成的 RSA KeyID: %s", keyID)

	// 测试签名
	msg := []byte("RSA 签名消息")
	sig, err := kms.Sign(keyID, msg)
	if err != nil {
		t.Fatalf("RSA 签名失败: %v", err)
	}

	// 测试验证
	err = kms.Verify(keyID, msg, sig)
	if err != nil {
		t.Errorf("RSA 验签失败: %v", err)
	}

	// 测试加密 (RSA OAEP)
	plaintext := []byte("RSA 加密消息")
	ciphertext, err := kms.Encrypt(keyID, plaintext)
	if err != nil {
		t.Fatalf("RSA 加密失败: %v", err)
	}

	// 测试解密
	decrypted, err := kms.Decrypt(keyID, ciphertext)
	if err != nil {
		t.Fatalf("RSA 解密失败: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("RSA 解密内容不匹配")
	}
}

func TestLocalKMS_ECDSA(t *testing.T) {
	kms := NewLocalKMS()

	keyID, err := kms.CreateKey(KeyTypeECDSA)
	if err != nil {
		t.Fatalf("创建 ECDSA 密钥失败: %v", err)
	}
	t.Logf("生成的 ECDSA KeyID: %s", keyID)

	msg := []byte("ECDSA 签名消息")
	sig, err := kms.Sign(keyID, msg)
	if err != nil {
		t.Fatalf("ECDSA 签名失败: %v", err)
	}

	err = kms.Verify(keyID, msg, sig)
	if err != nil {
		t.Errorf("ECDSA 验签失败: %v", err)
	}
}

func TestLocalKMS_Ed25519(t *testing.T) {
	kms := NewLocalKMS()

	keyID, err := kms.CreateKey(KeyTypeEd25519)
	if err != nil {
		t.Fatalf("创建 Ed25519 密钥失败: %v", err)
	}
	t.Logf("生成的 Ed25519 KeyID: %s", keyID)

	msg := []byte("Ed25519 签名消息")
	sig, err := kms.Sign(keyID, msg)
	if err != nil {
		t.Fatalf("Ed25519 签名失败: %v", err)
	}

	err = kms.Verify(keyID, msg, sig)
	if err != nil {
		t.Errorf("Ed25519 验签失败: %v", err)
	}
}

func TestLocalKMS_Errors(t *testing.T) {
	kms := NewLocalKMS()

	// 测试不存在的 Key
	_, err := kms.Encrypt("non-existent", []byte("data"))
	if err != ErrKeyNotFound {
		t.Errorf("期望 ErrKeyNotFound, 得到 %v", err)
	}

	// 测试 AES 不支持签名
	aesKeyID, _ := kms.CreateKey(KeyTypeAES)
	_, err = kms.Sign(aesKeyID, []byte("data"))
	if err != ErrUnsupportedOp {
		t.Errorf("期望 ErrUnsupportedOp (AES 签名), 得到 %v", err)
	}
}
