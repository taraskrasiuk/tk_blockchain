package cmd

import (
	"fmt"
	"os"
	"taraskrasiuk/blockchain_l/internal/state"

	"github.com/spf13/cobra"
)

func addBalancesListCmd() *cobra.Command {
	var balancesListCmd = &cobra.Command{
		Use:   "list",
		Short: "Balances",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				dirname, _ = cmd.Flags().GetString("dir")
			)
			s := state.NewState(dirname)
			defer s.Close()

			res := fmt.Sprintf("Account balances at: %s\n", s.GetVersion())
			for acc, val := range s.Balances {
				res += "-----\n"
				res += fmt.Sprintf("%s : %d\n", acc, val)
				res += "-----\n"
			}
			fmt.Fprintf(os.Stdout, res)
		},
	}

	addRequiredArg(balancesListCmd)

	return balancesListCmd
}
