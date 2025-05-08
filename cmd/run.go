package cmd

import (
	"log"
	"taraskrasiuk/blockchain_l/internal/node"

	"github.com/spf13/cobra"
)

func addRunCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Run the HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				datadir, _ = cmd.Flags().GetString("dir")
			)
			err := node.Run(datadir)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	addRequiredArg(cmd)
	return cmd
}
