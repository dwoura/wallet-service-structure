package bip32

import (
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
)

// KeyType 定义密钥类型（私钥或公钥）
type KeyType int

const (
	PrivateKey KeyType = iota
	PublicKey
)

// ExtendedKey 包装了 BIP-32 扩展密钥
type ExtendedKey interface {
	// String 返回 Base58 编码的密钥字符串 (xprv... / xpub...)
	String() string

	// ECPubKey 用于获取底层的 EC 公钥
	ECPubKey() (*btcec.PublicKey, error)
	// ECPrivKey 用于获取底层的 EC 私钥 (用于签名)
	ECPrivKey() (*btcec.PrivateKey, error)
	// Derive 根据索引派生子密钥
	Derive(index uint32) (ExtendedKey, error)
	// IsPrivate 返回是否包含私钥
	IsPrivate() bool
	// Address 返回关联的地址 (默认 BTC 格式，具体取决于实现)
	Address() string
	// Neuter 返回对应的扩展公钥 (如果当前是私钥)
	Neuter() (ExtendedKey, error)
}

// HDWallet 定义了分层确定性钱包的基本行为
type HDWallet interface {
	// MasterKey 返回主扩展密钥
	MasterKey() ExtendedKey
	// DerivePath 根据路径 (如 "m/44'/0'/0'/0/0") 派生密钥
	DerivePath(path string) (ExtendedKey, error)
}

var (
	ErrInvalidSeed = errors.New("无效的种子")
	ErrInvalidPath = errors.New("无效的派生路径")
)
