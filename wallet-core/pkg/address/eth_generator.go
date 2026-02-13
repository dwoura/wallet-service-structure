package address

import (
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/sha3"
)

// ETHGenerator 以太坊地址生成器
type ETHGenerator struct{}

func NewETHGenerator() *ETHGenerator {
	return &ETHGenerator{}
}

// PubKeyToAddress 将公钥字节 (非压缩格式, 65 bytes, 0x04...) 转换为 EIP-55 地址
func (g *ETHGenerator) PubKeyToAddress(pubKeyBytes []byte) (string, error) {
	// 1. 去掉前缀 0x04 (如果存在)
	if len(pubKeyBytes) == 65 && pubKeyBytes[0] == 0x04 {
		pubKeyBytes = pubKeyBytes[1:]
	}

	// 2. Keccak-256 哈希
	hash := keccak256(pubKeyBytes)

	// 3. 取后 20 字节
	addressBytes := hash[12:]

	// 4. Hex 编码并添加 EIP-55 校验和
	addressHex := hex.EncodeToString(addressBytes)
	return "0x" + toChecksumAddress(addressHex), nil
}

func keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// toChecksumAddress 实现 EIP-55 混合大小写校验
func toChecksumAddress(address string) string {
	address = strings.ToLower(address)
	hash := keccak256([]byte(address))
	hexHash := hex.EncodeToString(hash)

	var sb strings.Builder
	for i := 0; i < len(address); i++ {
		char := address[i]
		// 检查 hash 的第 i 位是否 >= 8
		hashByte := hexCharToInt(hexHash[i])
		if hashByte >= 8 {
			sb.WriteString(strings.ToUpper(string(char)))
		} else {
			sb.WriteByte(char)
		}
	}
	return sb.String()
}

func hexCharToInt(c byte) byte {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	if c >= 'a' && c <= 'f' {
		return c - 'a' + 10
	}
	return 0
}
