package crypto_util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptAESGCM 使用给定的密钥对明文进行 AES-GCM 加密。
// 密钥必须是 16、24 或 32 字节长，分别对应 AES-128、AES-192 或 AES-256。
// 返回 nonce + 密文。
func EncryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
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

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// DecryptAESGCM 使用给定的密钥对 AES-GCM 密文（nonce + 加密数据）进行解密。
func DecryptAESGCM(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("密文太短")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ------------------------------------------------------------------------------------------------
// 警告：DES 不安全，严禁在生产环境的新系统中使用。
// 本实现仅用于教学目的，演示过时的算法。
// ------------------------------------------------------------------------------------------------

// EncryptDES 使用 DES 在 CBC 模式下加密明文。
// 密钥必须是 8 字节。IV 是随机生成的。
// 返回 iv + 密文。
func EncryptDES(key, plaintext []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("DES 密钥长度必须是 8 字节")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// PKCS7 填充
	padding := des.BlockSize - len(plaintext)%des.BlockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	plaintext = append(plaintext, padtext...)

	ciphertext := make([]byte, des.BlockSize+len(plaintext))
	iv := ciphertext[:des.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[des.BlockSize:], plaintext)

	return ciphertext, nil
}

// DecryptDES 解密 DES-CBC 加密的数据。
func DecryptDES(key, ciphertext []byte) ([]byte, error) {
	if len(key) != 8 {
		return nil, errors.New("DES 密钥长度必须是 8 字节")
	}
	if len(ciphertext) < des.BlockSize {
		return nil, errors.New("密文太短")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:des.BlockSize]
	ciphertext = ciphertext[des.BlockSize:]

	if len(ciphertext)%des.BlockSize != 0 {
		return nil, errors.New("密文不是块大小的倍数")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// PKCS7 去除填充
	length := len(plaintext)
	unpadding := int(plaintext[length-1])
	if unpadding > length {
		return nil, errors.New("去除填充错误")
	}
	return plaintext[:length-unpadding], nil
}
