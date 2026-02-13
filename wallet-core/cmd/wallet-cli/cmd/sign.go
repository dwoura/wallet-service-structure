package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"syscall"

	"wallet-core/pkg/bip32"
	"wallet-core/pkg/bip39"
	"wallet-core/pkg/keystore"
	"wallet-core/pkg/wallet/types"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types" // Alias to avoid conflict
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "离线签名交易 (Offline Signing)",
	Long:  `读取未签名的交易 JSON 文件，使用 Keystore 进行签名，并输出已签名的交易 (Raw Tx)。`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFile, _ := cmd.Flags().GetString("input")
		outputFile, _ := cmd.Flags().GetString("output")
		keystoreFile, _ := cmd.Flags().GetString("keystore")

		// 1. 读取未签名交易
		data, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("读取输入文件失败: %v\n", err)
			os.Exit(1)
		}

		var unsignedTx types.UnsignedTransaction
		if err := json.Unmarshal(data, &unsignedTx); err != nil {
			fmt.Printf("解析交易文件失败: %v\n", err)
			os.Exit(1)
		}

		// 显示交易详情供用户确认 (Verify on Screen)
		fmt.Println("\n================ 待签名交易 ================")
		fmt.Printf("Chain:      %s (ID: %d)\n", unsignedTx.Chain, unsignedTx.ChainID)
		fmt.Printf("From:       %s\n", unsignedTx.From)
		fmt.Printf("To:         %s\n", unsignedTx.To)
		fmt.Printf("Amount:     %s\n", unsignedTx.Amount)
		fmt.Printf("Nonce:      %d\n", unsignedTx.Nonce)
		fmt.Printf("GasPrice:   %s\n", unsignedTx.GasPrice)
		fmt.Printf("Path:       %s\n", unsignedTx.DerivationPath)
		fmt.Println("============================================")

		// 2. 加载 Keystore
		fmt.Printf("\n正在从 %s 加载 Keystore...\n", keystoreFile)
		encryptedKey, err := keystore.LoadFromFile(keystoreFile)
		if err != nil {
			fmt.Printf("加载 Keystore 失败: %v\n", err)
			os.Exit(1)
		}

		// 3. 输入密码并解密
		fmt.Print("请输入 Keystore 密码以确认签名: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("\n读取密码失败:", err)
			os.Exit(1)
		}
		fmt.Println()
		password := string(bytePassword)

		mnemonic, err := keystore.DecryptMnemonic(encryptedKey, password)
		if err != nil {
			fmt.Printf("解密失败 (密码错误?): %v\n", err)
			os.Exit(1)
		}

		// 4. 恢复 Master Key 并派生私钥
		seed := bip39.NewMnemonicService().MnemonicToSeed(mnemonic, "")

		wallet, err := bip32.NewMasterKeyFromSeed(seed, &chaincfg.MainNetParams)
		if err != nil {
			fmt.Printf("恢复 Master Key 失败: %v\n", err)
			os.Exit(1)
		}

		// 5. 派生私钥 using DerivePath
		derivedKey, err := wallet.DerivePath(unsignedTx.DerivationPath)
		if err != nil {
			fmt.Printf("私钥派生失败: %v\n", err)
			os.Exit(1)
		}

		// 6. 签名 (目前只支持 ETH)
		if unsignedTx.Chain != "ETH" {
			fmt.Println("目前仅支持 ETH 签名")
			os.Exit(1)
		}

		rawTxHex, txHash, err := signEthTx(derivedKey, &unsignedTx)
		if err != nil {
			fmt.Printf("签名失败: %v\n", err)
			os.Exit(1)
		}

		// 7. 输出结果
		signedTx := types.SignedTransaction{
			TxHash: txHash,
			RawTx:  rawTxHex,
		}

		outputData, _ := json.MarshalIndent(signedTx, "", "  ")
		if err := os.WriteFile(outputFile, outputData, 0644); err != nil {
			fmt.Printf("保存结果失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ 签名成功!\n")
		fmt.Printf("TxHash: %s\n", txHash)
		fmt.Printf("已保存到: %s\n", outputFile)
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringP("input", "i", "unsigned.json", "未签名的交易文件路径")
	signCmd.Flags().StringP("output", "o", "signed.json", "签名后的输出文件路径")
	signCmd.Flags().StringP("keystore", "k", "wallet.json", "Keystore 文件路径")
}

func signEthTx(privKey bip32.ExtendedKey, utx *types.UnsignedTransaction) (string, string, error) {
	// Convert types
	nonce := utx.Nonce

	amountInt, _ := new(big.Int).SetString(utx.Amount, 10)
	gasPriceInt, _ := new(big.Int).SetString(utx.GasPrice, 10)

	toAddr := common.HexToAddress(utx.To)

	// New Transaction
	var tx *ethtypes.Transaction
	if len(utx.Data) > 0 {
		data := common.FromHex(utx.Data)
		tx = ethtypes.NewTransaction(nonce, toAddr, amountInt, utx.GasLimit, gasPriceInt, data)
	} else {
		tx = ethtypes.NewTransaction(nonce, toAddr, amountInt, utx.GasLimit, gasPriceInt, nil)
	}

	// Sign
	pk, err := privKey.ECPrivKey()
	if err != nil {
		return "", "", err
	}
	ecdsaKey := pk.ToECDSA()

	chainID := big.NewInt(utx.ChainID)
	signer := ethtypes.NewEIP155Signer(chainID)

	signedTx, err := ethtypes.SignTx(tx, signer, ecdsaKey)
	if err != nil {
		return "", "", err
	}

	// Serialize (RLP Encoding)
	rawTxBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return "", "", err
	}
	rawTxHex := hex.EncodeToString(rawTxBytes)

	return "0x" + rawTxHex, signedTx.Hash().Hex(), nil
}
