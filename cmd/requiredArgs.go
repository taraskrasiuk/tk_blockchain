package cmd

import "github.com/spf13/cobra"

var (
	reqDirFlag = "dir"
)

func addRequiredArg(cmd *cobra.Command) {
	cmd.Flags().String(reqDirFlag, "", "the database directory")
	cmd.MarkFlagRequired(reqDirFlag)
}
