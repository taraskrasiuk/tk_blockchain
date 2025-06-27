package cmd

import (
	"context"
	"fmt"
	"log"
	"taraskrasiuk/blockchain_l/internal/database"
	"taraskrasiuk/blockchain_l/internal/node"
	"time"

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
			fromAcc := database.NewAccount(from)
			toAcc := database.NewAccount(to)

			tx := database.NewTx(fromAcc, toAcc, data, value)
			signedTx := database.NewSignedTx(*tx, []byte{})
			s, err := database.NewState(dirname, true)
			if err != nil {
				log.Fatal(err)
				return
			}
			defer s.Close()

			pendingBlock := node.NewPendingBlock(*s.GetLastHash(), s.NextBlockNumber(), []database.SignedTx{*signedTx}, database.NewAccount("miner"))
			miningCtx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
			defer cancel()
			block, err := node.Mine(miningCtx, pendingBlock)
			// bl := database.NewBlock(*s.GetLastHash(), s.NextBlockNumber(), nonce uint32, payload []database.SignedTx, miner database.Account)
			// TODO: probably need to change the method to apply the pointer
			// in order to avoid copying the values
			if _, err := s.AddBlock(block); err != nil {
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

	addRequiredArg(cmd)

	txCmd.AddCommand(cmd)

	return txCmd
}
