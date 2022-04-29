package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	connectionCommand.AddCommand(connectionListCommand)
}

var connectionListCommand = &cobra.Command{
	Use:   "list [[SERVICE NAME]]",
	Short: "List connections",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listResources("connection", args)
	},
}
