package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/spf13/cobra"

	"wallet-core/pkg/address"
	"wallet-core/pkg/bip32"
	"wallet-core/pkg/bip39"
)

// newCmd 代表 new 命令
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "创建一个新的钱包",
	Long:  `生成一个新的随机 BIP-39 助记词，并显示派生的种子和主密钥。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("正在生成新钱包...")
		fmt.Println("---------------------------------------------------")

		// 1. 生成助记词
		mnemonicService := bip39.NewMnemonicService()
		mnemonic, err := mnemonicService.GenerateMnemonic(256) // 24 words
		if err != nil {
			fmt.Printf("生成助记词失败: %v\n", err)
			return
		}
		fmt.Printf("助记词 (Mnemonic): \n%s\n", mnemonic)
		fmt.Println("---------------------------------------------------")

		// 2. 生成种子
		seed := mnemonicService.MnemonicToSeed(mnemonic, "")
		fmt.Printf("种子 (Seed Hex): %s\n", hex.EncodeToString(seed))

		// 3. 生成 HD Wallet 主密钥
		wallet, err := bip32.NewMasterKeyFromSeed(seed, &chaincfg.MainNetParams)
		if err != nil {
			fmt.Printf("生成主密钥失败: %v\n", err)
			return
		}
		masterKey := wallet.MasterKey()
		fmt.Printf("主私钥 (xprv): %s\n", masterKey.String())

		pubMasterKey, _ := masterKey.Neuter()
		fmt.Printf("主公钥 (xpub): %s\n", pubMasterKey.String())
		fmt.Println("---------------------------------------------------")

		// 4. 派生默认地址 (BIP-44)
		// Bitcoin: m/44'/0'/0'/0/0
		btcPath := "m/44'/0'/0'/0/0"
		btcKey, err := wallet.DerivePath(btcPath)
		if err != nil {
			fmt.Printf("BTC 派生失败: %v\n", err)
		} else {
			// 获取公钥用于生成地址
			btcPubKey, _ := btcKey.Neuter()
			// 注意: btcutil/hdkeychain 的 ExtendedKey 通常不能直接拿到原始 bytes，需要通过 ECPubKey()
			ecPubKey, err := btcPubKey.(*bip32.BTCKeychain).ECPubKey()
			if err == nil {
				btcGen := address.NewBTCGenerator(&chaincfg.MainNetParams)
				btcAddr, err := btcGen.PubKeyToAddress(ecPubKey.SerializeCompressed())
				if err == nil {
					fmt.Printf("Bitcoin Address (Mainnet) [%s]: %s\n", btcPath, btcAddr)
				}
			}
		}

		// Ethereum: m/44'/60'/0'/0/0
		ethPath := "m/44'/60'/0'/0/0"
		ethKey, err := wallet.DerivePath(ethPath)
		if err != nil {
			fmt.Printf("ETH 派生失败: %v\n", err)
		} else {
			ethPubKey, _ := ethKey.Neuter()
			ecPubKey, err := ethPubKey.(*bip32.BTCKeychain).ECPubKey()
			if err == nil {
				ethGen := address.NewETHGenerator()
				ethAddr, err := ethGen.PubKeyToAddress(ecPubKey.SerializeUncompressed())
				if err == nil {
					fmt.Printf("Ethereum Address [%s]: %s\n", ethPath, ethAddr)
				}
			}
		}
		fmt.Println("---------------------------------------------------")
		fmt.Println("请妥善保管您的助记词！任何拥有助记词的人都可以控制该钱包的所有资产。")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
