package cmd

import (
	"fmt"
	"log"
	"taraskrasiuk/blockchain_l/internal/node"

	"github.com/spf13/cobra"
)

func addRunCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Run the HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			datadir, _ := cmd.Flags().GetString("dir")
			port, err := cmd.Flags().GetUint("port")
			if err != nil {
				log.Fatal(err)
				return
			}
			// Check if running the bootstrap node or not
			isBootstrap, _ := cmd.Flags().GetBool("bootstrap")

			if isBootstrap {
				n := node.NewNode(datadir, port, nil)
				if err := n.Run(cmd.Context()); err != nil {
					log.Fatal(err)
					return
				}
			} else {
				bootstrapIp, err := cmd.Flags().GetString("bootstrapIp")
				if err != nil {
					log.Fatal(err)
					return
				}
				bootstrapPort, err := cmd.Flags().GetUint("bootstrapPort")
				if err != nil {
					log.Fatal(err)
					return
				}
				// set an ip of extrenal node
				boostrap := node.NewPeerNode(bootstrapIp, bootstrapPort, true, true)

				fmt.Printf("successfully added the bootstrap node with ip %s and port %d \n", bootstrapIp, bootstrapPort)

				n := node.NewNode(datadir, port, boostrap)
				if err := n.Run(cmd.Context()); err != nil {
					log.Fatal(err)
					return
				}
			}
		},
	}

	addRequiredArg(cmd)
	cmd.Flags().Uint("port", 8080, "Define the port number")
	cmd.MarkFlagRequired("port")
	cmd.Flags().Bool("bootstrap", false, "Is running a bootstrap node or not")
	cmd.Flags().String("bootstrapIp", "", "The ip of the bootstrap node")
	cmd.Flags().Uint("bootstrapPort", 8080, "The bootstrap node port")
	return cmd
}
