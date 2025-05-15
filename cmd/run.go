package cmd

import (
	"fmt"
	"log"
	"taraskrasiuk/blockchain_l/internal/node"
	"taraskrasiuk/blockchain_l/internal/server"

	"github.com/spf13/cobra"
)

const (
	DEFAULT_PORT              = 8080
	BOOTSTRAP_NODE_BY_DEFAULT = false
)

func runBootstrapNode(cmd *cobra.Command, datadir string, port uint) error {
	srv := server.NewNodeServer(node.NewNode(datadir, port, nil, true), port)

	return srv.Run(cmd.Context())
}

func runPeerNode(cmd *cobra.Command, datadir string, port uint) error {
	bootstrapIp, err := cmd.Flags().GetString("bootstrapIp")
	if err != nil {
		return err
	}
	bootstrapPort, err := cmd.Flags().GetUint("bootstrapPort")
	if err != nil {
		return err
	}
	// set an ip of extrenal node
	boostrap := node.NewPeerNode(bootstrapIp, bootstrapPort, true, true)
	fmt.Printf("successfully added the bootstrap node with ip %s and port %d \n", bootstrapIp, bootstrapPort)

	srv := server.NewNodeServer(node.NewNode(datadir, port, boostrap, true), port)
	if err := srv.Run(cmd.Context()); err != nil {
		return err
	}
	return nil
}

func addRunCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Run the HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				datadir, _     = cmd.Flags().GetString("dir")
				port, _        = cmd.Flags().GetUint("port")
				isBootstrap, _ = cmd.Flags().GetBool("bootstrap")
			)

			if isBootstrap {
				if err := runBootstrapNode(cmd, datadir, port); err != nil {
					log.Fatal(err)
				}
			} else {
				if err := runPeerNode(cmd, datadir, port); err != nil {
					log.Fatal(err)
				}
			}
		},
	}

	addRequiredArg(cmd)
	cmd.Flags().Uint("port", DEFAULT_PORT, "Define the port number")
	cmd.MarkFlagRequired("port")
	cmd.Flags().Bool("bootstrap", BOOTSTRAP_NODE_BY_DEFAULT, "Is running a bootstrap node or not")
	cmd.Flags().String("bootstrapIp", "", "The ip of the bootstrap node")
	cmd.Flags().Uint("bootstrapPort", DEFAULT_PORT, "The bootstrap node port")
	return cmd
}
