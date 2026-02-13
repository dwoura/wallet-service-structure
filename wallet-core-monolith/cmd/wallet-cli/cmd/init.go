package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"wallet-core/pkg/bip39"
	"wallet-core/pkg/keystore"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化一个新的钱包 (生成助记词并加密保存)",
	Long:  `生成新的 BIP-39 助记词，并使用用户输入的密码进行加密，保存为 wallet.json 文件。`,
	Run: func(cmd *cobra.Command, args []string) {
		outputFile, _ := cmd.Flags().GetString("output")
		if _, err := os.Stat(outputFile); err == nil {
			fmt.Printf("错误: 文件 %s 已存在。请先删除或指定其他文件名。\n", outputFile)
			os.Exit(1)
		}

		fmt.Println("正在初始化新钱包...")
		fmt.Println("请设置一个强密码来保护您的助记词。")

		// 1. 输入密码
		fmt.Print("输入密码: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("\n读取密码失败:", err)
			os.Exit(1)
		}
		password := string(bytePassword)
		fmt.Println()

		fmt.Print("确认密码: ")
		bytePasswordConfirm, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Println("\n读取密码失败:", err)
			os.Exit(1)
		}
		confirmPassword := string(bytePasswordConfirm)
		fmt.Println()

		if password != confirmPassword {
			fmt.Println("两次输入的密码不一致！")
			os.Exit(1)
		}

		if len(password) < 6 {
			fmt.Println("密码长度至少需要 6 位。")
			os.Exit(1)
		}

		// 2. 生成助记词
		fmt.Println("正在生成助记词...")
		service := bip39.NewMnemonicService()
		mnemonic, err := service.GenerateMnemonic(12) // 默认 12 词
		if err != nil {
			fmt.Printf("生成助记词失败: %v\n", err)
			os.Exit(1)
		}

		// 3. 加密
		fmt.Println("正在加密保存...")
		encryptedKey, err := keystore.EncryptMnemonic(mnemonic, password)
		if err != nil {
			fmt.Printf("加密失败: %v\n", err)
			os.Exit(1)
		}

		// 4. 保存
		err = encryptedKey.SaveToFile(outputFile)
		if err != nil {
			fmt.Printf("保存文件失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ 钱包已初始化！\n")
		fmt.Printf("文件位置: %s\n", outputFile)
		fmt.Printf("您的 ID: %s\n", encryptedKey.Id)
		fmt.Println("\n⚠️  警告: 请务必记住您的密码！如果丢失密码，您将无法恢复钱包。")

		// 询问是否显示助记词 (Optional security risk, but good for learning/backup)
		fmt.Print("\n是否需要现在显示助记词以便备份? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			fmt.Println("\n---------------------------------------------------")
			fmt.Println("助记词 (请抄写在纸上并安全保管):")
			fmt.Println(mnemonic)
			fmt.Println("---------------------------------------------------")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("output", "o", "wallet.json", "输出的 Keystore 文件名")
}
