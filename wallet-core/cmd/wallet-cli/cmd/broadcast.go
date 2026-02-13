package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"wallet-core/pkg/wallet/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types" // Alias to avoid conflict
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

var broadcastCmd = &cobra.Command{
	Use:   "broadcast",
	Short: "广播已签名的交易 (Online)",
	Long:  `读取已签名的交易文件 (Signed Tx)，并广播到区块链网络。`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFile, _ := cmd.Flags().GetString("input")
		rpcURL, _ := cmd.Flags().GetString("rpc")

		// 1. 读取 Signed Tx
		data, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("读取文件失败: %v\n", err)
			os.Exit(1)
		}

		var signedTx types.SignedTransaction
		if err := json.Unmarshal(data, &signedTx); err != nil {
			fmt.Printf("解析文件失败: %v\n", err)
			os.Exit(1)
		}

		// 2. 连接节点
		fmt.Printf("正在连接 RPC: %s ...\n", rpcURL)
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			fmt.Printf("连接失败: %v\n", err)
			os.Exit(1)
		}

		// 3. 反序列化 Raw Tx
		rawTxBytes := common.FromHex(signedTx.RawTx)
		tx := new(ethtypes.Transaction)
		if err := tx.UnmarshalBinary(rawTxBytes); err != nil {
			fmt.Printf("反序列化交易失败: %v\n", err)
			os.Exit(1)
		}

		// 4. 广播
		fmt.Printf("正在广播交易 Hash: %s ...\n", tx.Hash().Hex())
		err = client.SendTransaction(context.Background(), tx)
		if err != nil {
			fmt.Printf("❌ 广播失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ 广播成功!\n")
		fmt.Printf("Tx URL: https://etherscan.io/tx/%s (Mainnet)\n", tx.Hash().Hex())
	},
}

func init() {
	rootCmd.AddCommand(broadcastCmd)
	broadcastCmd.Flags().StringP("input", "i", "signed.json", "已签名的交易文件")
	broadcastCmd.Flags().String("rpc", "https://cloudflare-eth.com", "RPC 节点地址")
}
