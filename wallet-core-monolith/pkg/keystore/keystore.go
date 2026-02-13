package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/scrypt"
)

// EncryptedKeyJSON 遵循 Ethereum Keystore V3 的结构风格
// 但为了简单和通用，我们存储的是 "助记词" (Mnemonic) 而不是单个私钥
type EncryptedKeyJSON struct {
	Crypto  CryptoJSON `json:"crypto"`
	Id      string     `json:"id"`      // UUID
	Version int        `json:"version"` // 3
}

type CryptoJSON struct {
	Cipher       string       `json:"cipher"`       // "aes-128-ctr" or "aes-256-gcm"
	CipherText   string       `json:"ciphertext"`   // Hex string
	CipherParams CipherParams `json:"cipherparams"` // IV
	KDF          string       `json:"kdf"`          // "scrypt"
	KDFParams    KDFParams    `json:"kdfparams"`
	MAC          string       `json:"mac"` // Hex string
}

type CipherParams struct {
	IV string `json:"iv"` // Hex string
}

type KDFParams struct {
	DKLen int    `json:"dklen"` // Derived Key Length (32)
	N     int    `json:"n"`     // Scrypt N (262144)
	R     int    `json:"r"`     // Scrypt r (8)
	P     int    `json:"p"`     // Scrypt p (1)
	Salt  string `json:"salt"`  // Hex string
}

const (
	scryptN     = 262144
	scryptR     = 8
	scryptP     = 1
	scryptDKLen = 32
)

// EncryptMnemonic 将助记词使用密码加密为 JSON 结构
func EncryptMnemonic(mnemonic, password string) (*EncryptedKeyJSON, error) {
	// 1. 生成随机 Salt
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// 2. 使用 Scrypt 派生密钥
	// derivedKey 长度 = 32 (用于 AES) + 32 (用于 MAC) ?
	// 为了标准 V3 兼容性，通常 DKLen=32，然后切分？
	// 这里我们简化：DKLen=32，直接用作 AES-GCM 的 Key。MAC 另外计算。
	derivedKey, err := scrypt.Key([]byte(password), salt, scryptN, scryptR, scryptP, scryptDKLen)
	if err != nil {
		return nil, err
	}

	// 3. 使用 AES-256-GCM 加密
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(mnemonic), nil)

	// 4. 计算 MAC (Verify Hash)
	// MAC = keccak256(derivedKey[16:32] + ciphertext) <- Eth V3 standard
	// 但我们要简单点，直接 SHA256(derivedKey + ciphertext)
	mac := sha256.Sum256(append(derivedKey, ciphertext...))

	// 5. 构造 JSON
	return &EncryptedKeyJSON{
		Version: 3,
		Id:      generateUUID(), // 简单 UUID 生成
		Crypto: CryptoJSON{
			Cipher:     "aes-256-gcm",
			CipherText: fmt.Sprintf("%x", ciphertext),
			CipherParams: CipherParams{
				IV: fmt.Sprintf("%x", nonce),
			},
			KDF: "scrypt",
			KDFParams: KDFParams{
				DKLen: scryptDKLen,
				N:     scryptN,
				R:     scryptR,
				P:     scryptP,
				Salt:  fmt.Sprintf("%x", salt),
			},
			MAC: fmt.Sprintf("%x", mac),
		},
	}, nil
}

// DecryptMnemonic 解密 Keystore JSON 获取助记词
func DecryptMnemonic(keyJSON *EncryptedKeyJSON, password string) (string, error) {
	// 1. 解析 Hex 参数
	salt, err := parseHex(keyJSON.Crypto.KDFParams.Salt)
	if err != nil {
		return "", fmt.Errorf("invalid salt: %v", err)
	}
	nonce, err := parseHex(keyJSON.Crypto.CipherParams.IV)
	if err != nil {
		return "", fmt.Errorf("invalid iv: %v", err)
	}
	ciphertext, err := parseHex(keyJSON.Crypto.CipherText)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %v", err)
	}
	mac, err := parseHex(keyJSON.Crypto.MAC)
	if err != nil {
		return "", fmt.Errorf("invalid mac: %v", err)
	}

	// 2. 重新派生密钥
	derivedKey, err := scrypt.Key([]byte(password), salt,
		keyJSON.Crypto.KDFParams.N,
		keyJSON.Crypto.KDFParams.R,
		keyJSON.Crypto.KDFParams.P,
		keyJSON.Crypto.KDFParams.DKLen)
	if err != nil {
		return "", err
	}

	// 3. 验证 MAC
	calculatedMAC := sha256.Sum256(append(derivedKey, ciphertext...))
	if !compareMAC(mac, calculatedMAC[:]) {
		return "", errors.New("invalid password or corrupted data (MAC mismatch)")
	}

	// 4. 解密
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}

	return string(plaintext), nil
}

// SaveToFile 保存到文件
func (k *EncryptedKeyJSON) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(k, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600) // 0600 is important
}

// LoadFromFile 从文件加载
func LoadFromFile(filename string) (*EncryptedKeyJSON, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var k EncryptedKeyJSON
	err = json.Unmarshal(data, &k)
	return &k, err
}

// --- Helpers ---

func parseHex(s string) ([]byte, error) {
	var b []byte
	_, err := fmt.Sscanf(s, "%x", &b)
	return b, err
}

func compareMAC(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func generateUUID() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
