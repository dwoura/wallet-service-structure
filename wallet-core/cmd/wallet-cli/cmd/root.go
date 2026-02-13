package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd 代表基础命令，没有子命令时直接调用
var rootCmd = &cobra.Command{
	Use:   "wallet-cli",
	Short: "区块链钱包命令行工具",
	Long: `一个用 Go 语言编写的区块链钱包学习工具。
支持生成 BIP-39 助记词、BIP-32 分层确定性钱包以及生成 BTC/ETH 地址。`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute 将所有子命令添加到根命令并设置标志
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 在这里可以定义全局标志 (Global Flags)
}
