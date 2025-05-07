package tx

import (
	"github.com/spf13/cobra"
)

var TxsCmd = &cobra.Command{
	Use:   "tx",
	Short: "Interact by transactions ( add )",
}

func init() {
	TxsCmd.AddCommand(txAddCmd())
}
