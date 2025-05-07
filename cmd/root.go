package cmd

import (
	"os"
	"taraskrasiuk/blockchain_l/cmd/tx"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "blockchain_l application",
	Short: "A brief description of your application",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(balancesListCmd)
	rootCmd.AddCommand(tx.TxsCmd)
}
