package crypto_util

import (
	"testing"
)

func TestRSA(t *testing.T) {
	// 生成密钥
	priv, pub, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair 失败: %v", err)
	}

	// 签名
	msg := []byte("Hello RSA")
	sig, err := RSASign(priv, msg)
	if err != nil {
		t.Fatalf("RSASign 失败: %v", err)
	}

	// 验证
	err = RSAVerify(pub, msg, sig)
	if err != nil {
		t.Errorf("RSAVerify 失败: %v", err)
	}

	// 验证失败用例
	msg2 := []byte("Hello RSA Modified")
	err = RSAVerify(pub, msg2, sig)
	if err == nil {
		t.Error("对于篡改后的消息，RSAVerify 应该失败")
	}
}

func TestECDSA(t *testing.T) {
	// 生成密钥
	priv, pub, err := GenerateECDSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateECDSAKeyPair 失败: %v", err)
	}

	// 签名
	msg := []byte("Hello ECDSA")
	sig, err := ECDSASign(priv, msg)
	if err != nil {
		t.Fatalf("ECDSASign 失败: %v", err)
	}

	// 验证
	valid := ECDSAVerify(pub, msg, sig)
	if !valid {
		t.Error("对于有效的签名，ECDSAVerify 返回了 false")
	}

	// 验证失败用例
	msg2 := []byte("Hello ECDSA Modified")
	valid = ECDSAVerify(pub, msg2, sig)
	if valid {
		t.Error("对于篡改后的消息，ECDSAVerify 应该返回 false")
	}
}

func TestEd25519(t *testing.T) {
	// 生成密钥
	priv, pub, err := GenerateEd25519KeyPair()
	if err != nil {
		t.Fatalf("GenerateEd25519KeyPair 失败: %v", err)
	}

	// 签名
	msg := []byte("Hello Ed25519")
	sig := Ed25519Sign(priv, msg)

	// 验证
	valid := Ed25519Verify(pub, msg, sig)
	if !valid {
		t.Error("对于有效的签名，Ed25519Verify 返回了 false")
	}

	// 验证失败用例
	msg2 := []byte("Hello Ed25519 Modified")
	valid = Ed25519Verify(pub, msg2, sig)
	if valid {
		t.Error("对于篡改后的消息，Ed25519Verify 应该返回 false")
	}
}

func TestPEMExport(t *testing.T) {
	priv, pub, err := GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatal(err)
	}

	pemPriv := ExportRSAPrivateKeyAsPEM(priv)
	if pemPriv == "" {
		t.Error("ExportRSAPrivateKeyAsPEM 返回为空")
	}

	pemPub, err := ExportRSAPublicKeyAsPEM(pub)
	if err != nil {
		t.Fatal(err)
	}
	if pemPub == "" {
		t.Error("ExportRSAPublicKeyAsPEM 返回为空")
	}
}
