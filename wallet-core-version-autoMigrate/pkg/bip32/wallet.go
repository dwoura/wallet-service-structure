package bip32

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

// BTCKeychain 实现了 ExtendedKey 接口，封装了 hdkeychain.ExtendedKey
type BTCKeychain struct {
	key     *hdkeychain.ExtendedKey
	network *chaincfg.Params
}

func (k *BTCKeychain) String() string {
	return k.key.String()
}

func (k *BTCKeychain) ECPubKey() (*btcec.PublicKey, error) {
	return k.key.ECPubKey()
}

// ECPrivKey 返回椭圆曲线私钥
func (k *BTCKeychain) ECPrivKey() (*btcec.PrivateKey, error) {
	return k.key.ECPrivKey()
}

func (k *BTCKeychain) Derive(index uint32) (ExtendedKey, error) {
	childKey, err := k.key.Derive(index)
	if err != nil {
		return nil, fmt.Errorf("派生子密钥失败: %v", err)
	}
	return &BTCKeychain{key: childKey, network: k.network}, nil
}

func (k *BTCKeychain) IsPrivate() bool {
	return k.key.IsPrivate()
}

func (k *BTCKeychain) Address() string {
	// 这是一个简化的实现，直接生成 P2PKH 地址
	addr, err := k.key.Address(k.network)
	if err != nil {
		return "unknown"
	}
	return addr.EncodeAddress()
}

func (k *BTCKeychain) Neuter() (ExtendedKey, error) {
	neuterKey, err := k.key.Neuter()
	if err != nil {
		return nil, fmt.Errorf("转换公钥失败: %v", err)
	}
	return &BTCKeychain{key: neuterKey, network: k.network}, nil
}

// Wallet 实现 HDWallet 接口
type Wallet struct {
	masterKey *BTCKeychain
	network   *chaincfg.Params
}

// NewMasterKeyFromSeed 使用 BIP-39 种子生成主密钥
// network: 默认为 chaincfg.MainNetParams
func NewMasterKeyFromSeed(seed []byte, network *chaincfg.Params) (*Wallet, error) {
	if len(seed) < 16 || len(seed) > 64 {
		return nil, ErrInvalidSeed
	}

	if network == nil {
		network = &chaincfg.MainNetParams
	}

	masterKey, err := hdkeychain.NewMaster(seed, network)
	if err != nil {
		return nil, fmt.Errorf("生成主密钥失败: %v", err)
	}

	return &Wallet{
		masterKey: &BTCKeychain{key: masterKey, network: network},
		network:   network,
	}, nil
}

func (w *Wallet) MasterKey() ExtendedKey {
	return w.masterKey
}

// DerivePath 解析路径并派生密钥
// 支持格式: m/44'/0'/0'/0/0 或 m/44h/0h/0h/0/0
func (w *Wallet) DerivePath(path string) (ExtendedKey, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return w.masterKey, nil
	}

	if strings.HasPrefix(path, "m/") {
		path = path[2:]
	}

	segments := strings.Split(path, "/")
	currentKey := w.masterKey

	for _, segment := range segments {
		var index uint32
		var err error

		isHardened := false
		if strings.HasSuffix(segment, "'") || strings.HasSuffix(segment, "h") {
			isHardened = true
			segment = segment[:len(segment)-1]
		}

		val, err := strconv.ParseUint(segment, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("无效的路径段 '%s': %v", segment, err)
		}
		index = uint32(val)

		if isHardened {
			index += hdkeychain.HardenedKeyStart
		}

		nextKey, err := currentKey.Derive(index)
		if err != nil {
			return nil, err
		}

		// 类型断言回 BTCKeychain 以便继续循环
		if k, ok := nextKey.(*BTCKeychain); ok {
			currentKey = k
		} else {
			return nil, fmt.Errorf("内部错误: 密钥类型不匹配")
		}
	}

	return currentKey, nil
}
