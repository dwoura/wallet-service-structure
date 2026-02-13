package cmd

import (
	"fmt"
	"os"

	"wallet-core/pkg/mpc"

	"github.com/spf13/cobra"
)

var (
	secretHex string
	parts     int
	threshold int
	shares    []string
)

func init() {
	// Add mpc command
	rootCmd.AddCommand(mpcCmd)

	// Subcommand: split
	mpcCmd.AddCommand(splitCmd)
	splitCmd.Flags().StringVarP(&secretHex, "secret", "s", "", "Private Key (Hex)")
	splitCmd.Flags().IntVarP(&parts, "parts", "n", 3, "Total number of shares (N)")
	splitCmd.Flags().IntVarP(&threshold, "threshold", "t", 2, "Threshold to recover (M)")
	splitCmd.MarkFlagRequired("secret")

	// Subcommand: recover
	mpcCmd.AddCommand(recoverCmd)
	recoverCmd.Flags().StringSliceVarP(&shares, "shares", "S", nil, "List of shares (comma separated)")
	recoverCmd.MarkFlagRequired("shares")
}

var mpcCmd = &cobra.Command{
	Use:   "mpc",
	Short: "MPC / Secret Sharing tools",
	Long:  `Utilities for Shamir's Secret Sharing (Split and Recover keys).`,
}

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Split a secret into N shares",
	Run: func(cmd *cobra.Command, args []string) {
		res, err := mpc.Split(secretHex, parts, threshold)
		if err != nil {
			fmt.Printf("Error splitting secret: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🔐 Secret split into %d shares (Threshold: %d):\n", parts, threshold)
		for i, share := range res {
			fmt.Printf("Share %d: %s\n", i+1, share)
		}
	},
}

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover a secret from M shares",
	Run: func(cmd *cobra.Command, args []string) {
		if len(shares) < 2 {
			fmt.Println("Error: At least 2 shares are required.")
			os.Exit(1)
		}

		recovered, err := mpc.Recover(shares)
		if err != nil {
			fmt.Printf("Error recovering secret: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🔑 Recovered Secret: %s\n", recovered)
	},
}
