package cmd

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"taraskrasiuk/blockchain_l/internal/database"

	"github.com/spf13/cobra"
)

func addNodeCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "node",
		Short: "get latest block number and hash of current node",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				dirname, _ = cmd.Flags().GetString("dir")
			)

			s, _ := database.NewState(dirname, true)
			defer s.Close()
			lastBlock := s.GetLastBlock()

			fmt.Printf("Latest block number: '%s'\nLatest block hash: '%s'\n",
				strconv.Itoa(int(lastBlock.Header.Number)),
				hex.EncodeToString([]byte(s.GetLastHash()[:])),
			)
		},
	}
	addRequiredArg(cmd)

	return cmd
}
