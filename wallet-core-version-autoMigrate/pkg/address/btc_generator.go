package address

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

// BTCGenerator 比特币地址生成器
type BTCGenerator struct {
	network *chaincfg.Params
}

func NewBTCGenerator(network *chaincfg.Params) *BTCGenerator {
	return &BTCGenerator{network: network}
}

// PubKeyToAddress 将公钥字节 (压缩格式) 转换为 P2PKH 地址
func (g *BTCGenerator) PubKeyToAddress(pubKeyBytes []byte) (string, error) {
	addr, err := btcutil.NewAddressPubKey(pubKeyBytes, g.network)
	if err != nil {
		return "", err
	}
	return addr.AddressPubKeyHash().EncodeAddress(), nil
}
