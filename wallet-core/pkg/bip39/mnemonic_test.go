package bip39

import (
	"encoding/hex"
	"testing"
)

func TestGenerateMnemonic(t *testing.T) {
	service := NewMnemonicService()

	// 测试 12 个单词 (128 bits)
	mnemonic12, err := service.GenerateMnemonic(128)
	if err != nil {
		t.Fatalf("生成 12 词助记词失败: %v", err)
	}
	t.Logf("12 词助记词: %s", mnemonic12)

	// 验证生成的助记词是否有效
	if !service.ValidateMnemonic(mnemonic12) {
		t.Errorf("生成的 12 词助记词无效")
	}

	// 测试 24 个单词 (256 bits)
	mnemonic24, err := service.GenerateMnemonic(256)
	if err != nil {
		t.Fatalf("生成 24 词助记词失败: %v", err)
	}
	t.Logf("24 词助记词: %s", mnemonic24)

	// 验证生成的助记词是否有效
	if !service.ValidateMnemonic(mnemonic24) {
		t.Errorf("生成的 24 词助记词无效")
	}
}

func TestMnemonicToSeed(t *testing.T) {
	service := NewMnemonicService()

	// 已知的测试向量 (Test Vector)
	// 助记词: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	// 密码: ""
	// 预期 Seed (Hex): "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	expectedSeedHex := "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"

	if !service.ValidateMnemonic(mnemonic) {
		t.Fatalf("测试向量助记词无效")
	}

	seed := service.MnemonicToSeed(mnemonic, "")
	seedHex := hex.EncodeToString(seed)

	if seedHex != expectedSeedHex {
		t.Errorf("Seed 生成不匹配。\n预期: %s\n实际: %s", expectedSeedHex, seedHex)
	} else {
		t.Logf("测试向量 Seed 匹配成功: %s", seedHex)
	}
}

func TestValidateMnemonic_Invalid(t *testing.T) {
	service := NewMnemonicService()

	invalidMnemonic := "hello world invalid mnemonic phrase designed to fail validation check"
	if service.ValidateMnemonic(invalidMnemonic) {
		t.Errorf("期望验证失败，但验证通过了")
	}
}
