package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"taraskrasiuk/blockchain_l/internal/database"

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
			s, err := database.NewState(dirname, true)
			if err != nil {
				log.Fatal(err)
				return
			}
			defer s.Close()

			res := fmt.Sprintf("Account balances at: %s\n", hex.EncodeToString([]byte(s.GetLastHash()[:])))
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
