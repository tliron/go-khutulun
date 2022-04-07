package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	hostCommand.AddCommand(hostAddCommand)
}

var hostListCommand = &cobra.Command{
	Use:   "list",
	Short: "List hosts in a cluster",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
