package kms

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"wallet-core/pkg/crypto_util"
	"wallet-core/pkg/safe_random"
)

// keyEntry 是内部存储结构，包含私钥（敏感数据）和元数据
type keyEntry struct {
	Metadata   KeyMetadata
	PrivateKey any // *rsa.PrivateKey, *ecdsa.PrivateKey, or ed25519.PrivateKey
	PublicKey  any // *rsa.PublicKey, *ecdsa.PublicKey, or ed25519.PublicKey
}

// LocalKMS 是 KeyManager 接口的本地内存实现。
// 它模拟了一个硬件安全模块 (HSM)，私钥存储在内存中，不直接暴露给外部。
type LocalKMS struct {
	mu   sync.RWMutex
	keys map[string]*keyEntry
}

// NewLocalKMS 创建一个新的 LocalKMS 实例。
func NewLocalKMS() *LocalKMS {
	return &LocalKMS{
		keys: make(map[string]*keyEntry),
	}
}

// CreateKey 创建一个新的密钥，并返回其 ID。
func (kms *LocalKMS) CreateKey(kType KeyType) (string, error) {
	kms.mu.Lock()
	defer kms.mu.Unlock()

	// 生成一个随机 Key ID
	keyID, err := safe_random.GenerateRandomHexString(16)
	if err != nil {
		return "", fmt.Errorf("生成 KeyID 失败: %w", err)
	}

	var priv any
	var pub any

	// 根据类型生成具体的密钥对
	switch kType {
	case KeyTypeAES:
		// AES 256
		keyBytes, err := safe_random.GenerateRandomBytes(32)
		if err != nil {
			return "", err
		}
		priv = keyBytes
		pub = nil // 对称密钥没有公钥

	case KeyTypeRSA:
		// RSA 2048
		rsaPriv, rsaPub, err := crypto_util.GenerateRSAKeyPair(2048)
		if err != nil {
			return "", err
		}
		priv = rsaPriv
		pub = rsaPub

	case KeyTypeECDSA:
		// ECDSA P-256
		ecdsaPriv, ecdsaPub, err := crypto_util.GenerateECDSAKeyPair()
		if err != nil {
			return "", err
		}
		priv = ecdsaPriv
		pub = ecdsaPub

	case KeyTypeEd25519:
		// Ed25519
		edPriv, edPub, err := crypto_util.GenerateEd25519KeyPair()
		if err != nil {
			return "", err
		}
		priv = edPriv
		pub = edPub

	default:
		return "", fmt.Errorf("不支持的密钥类型: %s", kType)
	}

	// 存储密钥
	entry := &keyEntry{
		Metadata: KeyMetadata{
			KeyID:     keyID,
			Type:      kType,
			CreatedAt: time.Now().Unix(),
			Enabled:   true,
		},
		PrivateKey: priv,
		PublicKey:  pub,
	}
	kms.keys[keyID] = entry

	return keyID, nil
}

// GetPublicKey 获取指定密钥 ID 的公钥。
func (kms *LocalKMS) GetPublicKey(keyID string) (any, error) {
	kms.mu.RLock()
	defer kms.mu.RUnlock()

	entry, exists := kms.keys[keyID]
	if !exists {
		return nil, ErrKeyNotFound
	}
	if !entry.Metadata.Enabled {
		return nil, ErrKeyDisabled
	}
	if entry.PublicKey == nil {
		return nil, fmt.Errorf("密钥 %s 没有公钥 (可能是对称密钥)", keyID)
	}

	return entry.PublicKey, nil
}

// Sign 使用指定的密钥对数据进行签名。
func (kms *LocalKMS) Sign(keyID string, data []byte) ([]byte, error) {
	kms.mu.RLock()
	defer kms.mu.RUnlock()

	entry, exists := kms.keys[keyID]
	if !exists {
		return nil, ErrKeyNotFound
	}
	if !entry.Metadata.Enabled {
		return nil, ErrKeyDisabled
	}

	switch k := entry.PrivateKey.(type) {
	case *rsa.PrivateKey:
		return crypto_util.RSASign(k, data)
	case *ecdsa.PrivateKey:
		return crypto_util.ECDSASign(k, data)
	case ed25519.PrivateKey:
		return crypto_util.Ed25519Sign(k, data), nil
	default:
		return nil, ErrUnsupportedOp
	}
}

// Verify 验证签名是否有效。
func (kms *LocalKMS) Verify(keyID string, data []byte, signature []byte) error {
	kms.mu.RLock()
	defer kms.mu.RUnlock()

	entry, exists := kms.keys[keyID]
	if !exists {
		return ErrKeyNotFound
	}
	if !entry.Metadata.Enabled {
		return ErrKeyDisabled
	}

	var valid bool
	switch k := entry.PublicKey.(type) {
	case *rsa.PublicKey:
		err := crypto_util.RSAVerify(k, data, signature)
		if err == nil {
			valid = true
		}
	case *ecdsa.PublicKey:
		valid = crypto_util.ECDSAVerify(k, data, signature)
	case ed25519.PublicKey:
		valid = crypto_util.Ed25519Verify(k, data, signature)
	default:
		return ErrUnsupportedOp
	}

	if !valid {
		return ErrInvalidSignature
	}
	return nil
}

// Encrypt 使用指定的密钥加密数据。
func (kms *LocalKMS) Encrypt(keyID string, plaintext []byte) ([]byte, error) {
	kms.mu.RLock()
	defer kms.mu.RUnlock()

	entry, exists := kms.keys[keyID]
	if !exists {
		return nil, ErrKeyNotFound
	}
	if !entry.Metadata.Enabled {
		return nil, ErrKeyDisabled
	}

	switch k := entry.PrivateKey.(type) {
	case []byte: // AES Key
		return crypto_util.EncryptAESGCM(k, plaintext)
	// RSA 加密 Demo (通常用公钥加密，私钥解密，但这里为了接口统一，我们假设这是服务端加密存储场景)
	// 但按照标准，Encrypt 应该是用公钥。
	// 这里我们做一个特殊处理：如果是 RSA，我们用公钥加密。
	// 注意：PrivateKey 字段存的是 PrivateKey 对象，它里面包含 PublicKey。
	case *rsa.PrivateKey:
		// 使用 OAEP 模式加密，这比 PKCS1v1.5 更安全
		// 我们直接用 entry.PublicKey。
		pub, ok := entry.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("RSA 密钥数据损坏")
		}
		// 使用 sha256 作为 hash
		return rsa.EncryptOAEP(sha256.New(), safe_random.Reader, pub, plaintext, nil)

	default:
		return nil, ErrUnsupportedOp
	}
}

// Decrypt 使用指定的密钥解密数据。
func (kms *LocalKMS) Decrypt(keyID string, ciphertext []byte) ([]byte, error) {
	kms.mu.RLock()
	defer kms.mu.RUnlock()

	entry, exists := kms.keys[keyID]
	if !exists {
		return nil, ErrKeyNotFound
	}
	if !entry.Metadata.Enabled {
		return nil, ErrKeyDisabled
	}

	switch k := entry.PrivateKey.(type) {
	case []byte: // AES Key
		return crypto_util.DecryptAESGCM(k, ciphertext)
	case *rsa.PrivateKey:
		// RSA 解密
		// 使用 sha256 作为 hash
		return rsa.DecryptOAEP(sha256.New(), safe_random.Reader, k, ciphertext, nil)
	default:
		return nil, ErrUnsupportedOp
	}
}
