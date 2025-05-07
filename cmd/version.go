package cmd 

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Major  = "0"
	Minor  = "0"
	Patch  = "1"
	Verbal = "TX Add & Balances list"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "A version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s.%s.%s-beta %s", Major, Minor, Patch, Verbal)
	},
}
