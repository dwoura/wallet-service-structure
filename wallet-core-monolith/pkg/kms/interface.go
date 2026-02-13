package kms

import (
	"errors"
)

// KeyType 定义了支持的密钥类型
type KeyType string

const (
	KeyTypeAES       KeyType = "AES"       // 对称加密密钥
	KeyTypeRSA       KeyType = "RSA"       // 非对称加密密钥 (用于数据传输/签名)
	KeyTypeECDSA     KeyType = "ECDSA"     // 椭圆曲线密钥 (P-256)
	KeyTypeEd25519   KeyType = "Ed25519"   // Ed25519 密钥 (Solana 等)
	KeyTypeSecp256k1 KeyType = "Secp256k1" // 比特币/以太坊专用 (后续实现)
)

// KeyMetadata 包含密钥的元数据，不包含敏感的私钥信息
type KeyMetadata struct {
	KeyID     string  `json:"key_id"`     // 密钥唯一标识符
	Type      KeyType `json:"type"`       // 密钥类型
	CreatedAt int64   `json:"created_at"` // 创建时间戳
	Enabled   bool    `json:"enabled"`    // 是否启用
}

// KeyManager 定义了密钥管理服务的核心行为。
// 这是一个抽象接口，允许我们后续替换为真实的 HSM (硬件安全模块) 或云端 KMS (如 AWS KMS)。
type KeyManager interface {
	// CreateKey 创建一个新的密钥，并返回其 ID。
	// 注意：私钥永远不会离开 KMS 的安全边界。
	CreateKey(kType KeyType) (string, error)

	// GetPublicKey 获取指定密钥 ID 的公钥。
	// 对于对称密钥 (AES)，此操作应返回错误或特定的占位符。
	GetPublicKey(keyID string) (any, error)

	// Sign 使用指定的密钥对数据进行签名。
	// 仅适用于非对称密钥 (RSA, ECDSA, Ed25519)。
	Sign(keyID string, data []byte) ([]byte, error)

	// Verify 验证签名是否有效。
	// 虽然可以通过 GetPublicKey 获取公钥在外部验证，但 KMS 通常也提供验证接口。
	Verify(keyID string, data []byte, signature []byte) error

	// Encrypt 使用指定的密钥加密数据。
	// 适用于 AES (对称) 或 RSA (非对称)。
	Encrypt(keyID string, plaintext []byte) ([]byte, error)

	// Decrypt 使用指定的密钥解密数据。
	Decrypt(keyID string, ciphertext []byte) ([]byte, error)
}

var (
	ErrKeyNotFound      = errors.New("密钥未找到")
	ErrKeyDisabled      = errors.New("密钥已禁用")
	ErrUnsupportedOp    = errors.New("该密钥类型不支持此操作")
	ErrInvalidSignature = errors.New("签名无效")
)
