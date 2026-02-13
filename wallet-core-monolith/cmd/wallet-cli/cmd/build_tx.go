package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"wallet-core/pkg/wallet/types"

	"github.com/spf13/cobra"
)

// buildTxCmd 模拟 Online 端的 "构造交易"
var buildTxCmd = &cobra.Command{
	Use:   "build-tx",
	Short: "构造未签名交易 (Online)",
	Long:  `模拟在线服务构造交易的过程。输入参数，输出 unsigned.json。`,
	Run: func(cmd *cobra.Command, args []string) {
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		amount, _ := cmd.Flags().GetString("amount")
		nonce, _ := cmd.Flags().GetUint64("nonce")
		path, _ := cmd.Flags().GetString("path")
		chainID, _ := cmd.Flags().GetInt64("chain-id")

		// 默认值
		gasLimit := uint64(21000)
		gasPrice := "20000000000" // 20 Gwei

		tx := types.UnsignedTransaction{
			Chain:          "ETH",
			From:           from,
			To:             to,
			Amount:         amount,
			Nonce:          nonce,
			GasLimit:       gasLimit,
			GasPrice:       gasPrice,
			Data:           "",
			DerivationPath: path,
			ChainID:        chainID,
		}

		outputFile, _ := cmd.Flags().GetString("output")
		data, _ := json.MarshalIndent(tx, "", "  ")

		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			fmt.Printf("保存失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 未签名交易已构造!\n文件: %s\n", outputFile)
	},
}

func init() {
	rootCmd.AddCommand(buildTxCmd)

	buildTxCmd.Flags().String("from", "", "发送方地址")
	buildTxCmd.Flags().String("to", "", "接收方地址")
	buildTxCmd.Flags().String("amount", "0", "金额 (Wei)")
	buildTxCmd.Flags().Uint64("nonce", 0, "Nonce")
	buildTxCmd.Flags().String("path", "m/44'/60'/0'/0/0", "私钥派生路径")
	buildTxCmd.Flags().Int64("chain-id", 1, "Chain ID (1=Mainnet, 11155111=Sepolia)")
	buildTxCmd.Flags().StringP("output", "o", "unsigned.json", "输出文件")

	buildTxCmd.MarkFlagRequired("from")
	buildTxCmd.MarkFlagRequired("to")
}
