package tx

import (
	"fmt"
	"log"
	"taraskrasiuk/blockchain_l/internal/state"
	"taraskrasiuk/blockchain_l/internal/transactions"

	"github.com/spf13/cobra"
)

func txAddCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add",
		Short: "Add transaction",
		PostRun: func(cmd *cobra.Command, args []string) {
			fmt.Println("Successfully transaction was added.")
		},
		Run: func(cmd *cobra.Command, args []string) {
			var (
				// ignore errors
				from, _  = cmd.Flags().GetString("from")
				to, _    = cmd.Flags().GetString("to")
				value, _ = cmd.Flags().GetUint("value")
				data, _  = cmd.Flags().GetString("data")
			)

			// create an accounts
			// create a transaction
			// get a state
			// add transaction to state
			// persist the state
			fromAcc := transactions.NewAccount(from)
			toAcc := transactions.NewAccount(to)

			tx := transactions.NewTx(fromAcc, toAcc, data, value)

			s := state.NewState()
			defer s.Close()

			// TODO: probably need to change the method to apply the pointer
			// in order to avoid copying the values
			err := s.Add(*tx)
			if err != nil {
				log.Fatal(err)
				return
			}
			// save updated state back to file
			err = s.Persist()
			if err != nil {
				log.Fatal(err)
				return
			}
		},
	}
	cmd.Flags().String("from", "", "from what account to perform the transaction")
	cmd.MarkFlagRequired("from")
	cmd.Flags().String("to", "", "to what account perform the transaction")
	cmd.MarkFlagRequired("to")
	cmd.Flags().Uint("value", 0, "amount of value, non negative")
	cmd.MarkFlagRequired("value")
	cmd.Flags().String("data", "", "data to send. Only 1 available option, is 'reward'.")

	return cmd
}
