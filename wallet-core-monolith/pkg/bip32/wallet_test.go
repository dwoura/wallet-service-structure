package bip32

import (
	"encoding/hex"
	"testing"

	"wallet-core/pkg/bip39"

	"github.com/btcsuite/btcd/chaincfg"
)

func TestNewMasterKeyFromSeed(t *testing.T) {
	// 使用 BIP-39 生成种子
	mnemonicService := bip39.NewMnemonicService()
	mnemonic, err := mnemonicService.GenerateMnemonic(128)
	if err != nil {
		t.Fatalf("生成助记词失败: %v", err)
	}
	seed := mnemonicService.MnemonicToSeed(mnemonic, "")

	// 生成主密钥
	wallet, err := NewMasterKeyFromSeed(seed, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("生成主密钥失败: %v", err)
	}

	if wallet.MasterKey() == nil {
		t.Fatalf("主密钥为空")
	}

	t.Logf("主私钥 (xprv): %s", wallet.MasterKey().String())
}

func TestDerivePath(t *testing.T) {
	// 测试向量
	// Seed: "fffcf9f6da3247d8a846f4b6113e6173" (16 bytes)
	seedHex := "fffcf9f6da3247d8a846f4b6113e6173"
	seed, _ := hex.DecodeString(seedHex)

	wallet, err := NewMasterKeyFromSeed(seed, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("生成主密钥失败: %v", err)
	}

	// 测试路径: m/0
	// 预期公钥 (xpub): xpub69H7F5d8KSRgmmdJg2KHPXV8a8qZq... (仅示例，这里不校验具体值，只校验过程)
	path1 := "m/0"
	child1, err := wallet.DerivePath(path1)
	if err != nil {
		t.Errorf("派生路径 %s 失败: %v", path1, err)
	}
	t.Logf("m/0 xprv: %s", child1.String())

	// 测试 Hardened 路径: m/0'
	path2 := "m/0'"
	child2, err := wallet.DerivePath(path2)
	if err != nil {
		t.Errorf("派生路径 %s 失败: %v", path2, err)
	}
	t.Logf("m/0' xprv: %s", child2.String())

	// 测试多层路径: m/44'/0'/0'/0/0 (BIP-44 BTC)
	path3 := "m/44'/0'/0'/0/0"
	child3, err := wallet.DerivePath(path3)
	if err != nil {
		t.Errorf("派生路径 %s 失败: %v", path3, err)
	}
	t.Logf("BIP-44 BTC xprv: %s", child3.String())

	// 验证公钥转换
	pubKey, err := child3.Neuter()
	if err != nil {
		t.Fatalf("转换为扩展公钥失败: %v", err)
	}
	t.Logf("BIP-44 BTC xpub: %s", pubKey.String())

	if pubKey.IsPrivate() {
		t.Errorf("Neuter() 应该返回公钥，但 IsPrivate() 返回 true")
	}
}
