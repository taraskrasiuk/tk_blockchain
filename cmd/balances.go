package cmd

import (
	"fmt"
	"os"
	"taraskrasiuk/blockchain_l/internal/state"

	"github.com/spf13/cobra"
)

var balancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Balances",
	Run: func(cmd *cobra.Command, args []string) {
		s := state.NewState()
		defer s.Close()

		res := "Balances: \n"
		for acc, val := range s.Balances {
			res += "-----\n"
			res += fmt.Sprintf("%s : %d\n", acc, val)
			res += "-----\n"
		}
		fmt.Fprintf(os.Stdout, res)
	},
}
