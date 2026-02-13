package bip39

import (
	"fmt"

	"github.com/tyler-smith/go-bip39"
)

// MnemonicService 提供助记词相关的功能
type MnemonicService struct{}

// NewMnemonicService 创建一个新的助记词服务实例
func NewMnemonicService() *MnemonicService {
	return &MnemonicService{}
}

// GenerateMnemonic 生成一个新的随机助记词 (BIP-39)。
// bitSize: 熵的位数，通常为 128 (12个单词) 或 256 (24个单词)。
func (s *MnemonicService) GenerateMnemonic(bitSize int) (string, error) {
	// 生成熵
	entropy, err := bip39.NewEntropy(bitSize)
	if err != nil {
		return "", fmt.Errorf("生成熵失败: %v", err)
	}

	// 从熵生成助记词
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("生成助记词失败: %v", err)
	}

	return mnemonic, nil
}

// ValidateMnemonic 验证助记词是否有效。
func (s *MnemonicService) ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// MnemonicToSeed 将助记词转换为种子 (BIP-39 Seed)。
// password: 可选的密码 (Passphrase),用于以此增强安全性 (这也是 "第25个单词" 的由来)。
// 如果不需要密码，传空字符串 ""。
func (s *MnemonicService) MnemonicToSeed(mnemonic string, password string) []byte {
	return bip39.NewSeed(mnemonic, password)
}
