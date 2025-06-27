package cmd

import (
	"fmt"
	"log"
	"taraskrasiuk/blockchain_l/internal/database"

	"github.com/spf13/cobra"
)

func addMigrationCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "migrate",
		Short: "migrate database",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				dirname, _ = cmd.Flags().GetString("dir")
			)
			s, _ := database.NewState(dirname, true)
			defer s.Close()

			block0 := database.NewBlock(database.Hash{}, 0, 0x0123, []database.SignedTx{
				*database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("andrej"), "", 3), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("andrej"), "reward", 700), []byte{}),
			}, database.NewAccount("miner"))

			block0Hash, err := s.AddBlock(block0)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("block hash: %x\n", block0Hash)
			fmt.Printf("parent block hash: %x\n", block0.Header.ParentHash)

			block1 := database.NewBlock(block0Hash, 1, 0x0123, []database.SignedTx{
				*database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("babayaga"), "", 2000), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("andrej"), "reward", 100), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("babayaga"), database.NewAccount("andrej"), "", 1), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("babayaga"), database.NewAccount("caesar"), "", 1000), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("babayaga"), database.NewAccount("andrej"), "", 50), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("andrej"), "reward", 600), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("andrej"), database.NewAccount("andrej"), "reward", 2600), []byte{}),
			}, database.NewAccount("miner"))

			block1Hash, err := s.AddBlock(block1)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("block hash: %x\n", block1Hash)
			fmt.Printf("parent block hash: %x\n", block1.Header.ParentHash)
		},
	}

	addRequiredArg(cmd)

	return cmd
}
