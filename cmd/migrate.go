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

			block0 := database.NewBlock(database.Hash{}, 0, 0x0123, []database.Tx{
				*database.NewTx(database.Account("andrej"), database.Account("andrej"), "", 3),
				*database.NewTx(database.Account("andrej"), database.Account("andrej"), "reward", 700),
			})

			s.AddBlock(block0)

			block0Hash, err := s.Persist()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("block hash: %x\n", block0Hash)
			fmt.Printf("parent block hash: %x\n", block0.Header.ParentHash)

			block1 := database.NewBlock(block0Hash, 1, 0x0123, []database.Tx{
				*database.NewTx("andrej", "babayaga", "", 2000),
				*database.NewTx("andrej", "andrej", "reward", 100),
				*database.NewTx("babayaga", "andrej", "", 1),
				*database.NewTx("babayaga", "caesar", "", 1000),
				*database.NewTx("babayaga", "andrej", "", 50),
				*database.NewTx("andrej", "andrej", "reward", 600),
				*database.NewTx("andrej", "andrej", "reward", 2600),
			})

			s.AddBlock(block1)

			block1Hash, err := s.Persist()
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
