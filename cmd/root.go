package cmd

import (
	"os"

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
	rootCmd.AddCommand(addBalancesListCmd())
	rootCmd.AddCommand(txAddCmd())
	rootCmd.AddCommand(addRunCmd())
	rootCmd.AddCommand(addMigrationCmd())
	rootCmd.AddCommand(addNodeCmd())
}
