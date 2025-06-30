package cmd

import (
	"fmt"
	"log"
	"taraskrasiuk/blockchain_l/internal/database"
	"taraskrasiuk/blockchain_l/internal/node"

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
			minerAcc := database.NewAccount("0xe3e1aff79a367675be32e76b63866ebd565f2b4a")
			n := node.NewNode(dirname, 8080, "localhost", nil, minerAcc, true)
			defer func() {
				if err := n.Close(); err != nil {
					log.Fatal(err)
				}
			}()

			pendingBlock0 := node.NewPendingBlock(database.Hash{}, 0, []database.SignedTx{
				*database.NewSignedTx(*database.NewTx(minerAcc, minerAcc, "", 3), []byte{}),
				*database.NewSignedTx(*database.NewTx(minerAcc, minerAcc, "reward", 700), []byte{}),
			}, minerAcc)
			block0, err := node.Mine(cmd.Context(), pendingBlock0)
			if err != nil {
				log.Fatal(err)
			}
			block0Hash, err := block0.Hash()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("block hash: %x\n", block0Hash)
			fmt.Printf("parent block hash: %x\n", block0.Header.ParentHash)

			pendingBlock1 := node.NewPendingBlock(block0Hash, 1, []database.SignedTx{
				*database.NewSignedTx(*database.NewTx(minerAcc, database.NewAccount("c9849c4f99c1a4a8fa57f0a6032f5e094acadeab"), "", 2000), []byte{}),
				*database.NewSignedTx(*database.NewTx(minerAcc, minerAcc, "reward", 100), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("c9849c4f99c1a4a8fa57f0a6032f5e094acadeab"), minerAcc, "", 1), []byte{}),
				*database.NewSignedTx(*database.NewTx(database.NewAccount("c9849c4f99c1a4a8fa57f0a6032f5e094acadeab"), minerAcc, "", 50), []byte{}),
				*database.NewSignedTx(*database.NewTx(minerAcc, minerAcc, "reward", 600), []byte{}),
				*database.NewSignedTx(*database.NewTx(minerAcc, minerAcc, "reward", 2600), []byte{}),
			}, minerAcc)

			block1, err := node.Mine(cmd.Context(), pendingBlock1)
			if err != nil {
				log.Fatal(err)
			}

			block1Hash, err := block1.Hash()
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
