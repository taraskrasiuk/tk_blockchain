package cmd

import (
	"fmt"
	"log"
	"taraskrasiuk/blockchain_l/internal/state"
	"taraskrasiuk/blockchain_l/internal/transactions"

	"github.com/spf13/cobra"
)

func txAddCmd() *cobra.Command {
	var txCmd = &cobra.Command{
		Use:   "tx",
		Short: "Transaction commands ( add )",
	}

	var cmd = &cobra.Command{
		Use:   "add",
		Short: "Add transaction",
		PostRun: func(cmd *cobra.Command, args []string) {
			fmt.Println("Successfully transaction was added.")
		},
		Run: func(cmd *cobra.Command, args []string) {
			var (
				// ignore errors
				dirname, _ = cmd.Flags().GetString("dir")
				from, _    = cmd.Flags().GetString("from")
				to, _      = cmd.Flags().GetString("to")
				value, _   = cmd.Flags().GetUint("value")
				data, _    = cmd.Flags().GetString("data")
			)

			// create an accounts
			// create a transaction
			// get a state
			// add transaction to state
			// persist the state
			fromAcc := transactions.NewAccount(from)
			toAcc := transactions.NewAccount(to)

			tx := transactions.NewTx(fromAcc, toAcc, data, value)

			s, err := state.NewState(dirname)
			if err != nil {
				log.Fatal(err)
				return
			}
			defer s.Close()

			// TODO: probably need to change the method to apply the pointer
			// in order to avoid copying the values
			if err := s.Add(*tx); err != nil {
				log.Fatal(err)
				return
			}
			// save updated state back to file
			snapshot, err := s.Persist()
			if err != nil {
				log.Fatal(err)
				return
			}
			fmt.Printf("Snapshot: %x\n", snapshot)
		},
	}
	cmd.Flags().String("from", "", "from what account to perform the transaction")
	cmd.MarkFlagRequired("from")
	cmd.Flags().String("to", "", "to what account perform the transaction")
	cmd.MarkFlagRequired("to")
	cmd.Flags().Uint("value", 0, "amount of value, non negative")
	cmd.MarkFlagRequired("value")
	cmd.Flags().String("data", "", "data to send. Only 1 available option, is 'reward'.")

	addRequiredArg(cmd)

	txCmd.AddCommand(cmd)

	return txCmd
}
