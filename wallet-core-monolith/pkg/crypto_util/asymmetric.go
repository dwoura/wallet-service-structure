package crypto_util

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
)

// ------------------------------------------------------------------------------------------------
// RSA (Rivest–Shamir–Adleman)
// ------------------------------------------------------------------------------------------------

// GenerateRSAKeyPair 生成指定位大小（例如 2048, 4096）的新 RSA 密钥对。
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return priv, &priv.PublicKey, nil
}

// RSASign 使用 PSS (概率签名方案) 对消息进行签名，该方案比 PKCS1v1.5 更安全。
func RSASign(priv *rsa.PrivateKey, message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)
	return rsa.SignPSS(rand.Reader, priv, crypto.SHA256, hash[:], nil)
}

// RSAVerify 验证 RSA-PSS 签名。
func RSAVerify(pub *rsa.PublicKey, message, signature []byte) error {
	hash := sha256.Sum256(message)
	return rsa.VerifyPSS(pub, crypto.SHA256, hash[:], signature, nil)
}

// ------------------------------------------------------------------------------------------------
// ECDSA (椭圆曲线数字签名算法) - 曲线 P-256
// 注意：比特币/以太坊使用 secp256k1，这需要外部库支持。P-256 是标准库自带的。
// ------------------------------------------------------------------------------------------------

// GenerateECDSAKeyPair 使用 Curve P-256 生成新的 ECDSA 密钥对。
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv, &priv.PublicKey, nil
}

// ECDSASign 对消息的哈希进行签名。返回 ASN.1 编码的签名。
func ECDSASign(priv *ecdsa.PrivateKey, message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)
	return ecdsa.SignASN1(rand.Reader, priv, hash[:])
}

// ECDSAVerify 验证 ASN.1 编码的 ECDSA 签名。
func ECDSAVerify(pub *ecdsa.PublicKey, message, signature []byte) bool {
	hash := sha256.Sum256(message)
	return ecdsa.VerifyASN1(pub, hash[:], signature)
}

// ------------------------------------------------------------------------------------------------
// Ed25519 (Edwards-curve 数字签名算法)
// 现代、快速且安全。是新一代非以太坊链（如 Solana 协议）的首选。
// ------------------------------------------------------------------------------------------------

// GenerateEd25519KeyPair 生成新的 Ed25519 密钥对。
func GenerateEd25519KeyPair() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	return priv, pub, err
}

// Ed25519Sign 对消息进行签名。
func Ed25519Sign(priv ed25519.PrivateKey, message []byte) []byte {
	return ed25519.Sign(priv, message)
}

// Ed25519Verify 验证签名。
func Ed25519Verify(pub ed25519.PublicKey, message, signature []byte) bool {
	return ed25519.Verify(pub, message, signature)
}

// ------------------------------------------------------------------------------------------------
// 导出/导入密钥的辅助函数 (PEM 格式) - 基础示例
// ------------------------------------------------------------------------------------------------

func ExportRSAPrivateKeyAsPEM(priv *rsa.PrivateKey) string {
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privBytes,
		},
	)
	return string(privPEM)
}

func ExportRSAPublicKeyAsPEM(pub *rsa.PublicKey) (string, error) {
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	pubPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubBytes,
		},
	)
	return string(pubPEM), nil
}
