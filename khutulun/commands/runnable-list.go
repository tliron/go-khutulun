package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	runnableCommand.AddCommand(runnableListCommand)
}

var runnableListCommand = &cobra.Command{
	Use:   "list [[SERVICE NAME]]",
	Short: "List runnables",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listResources("runnable", args)
	},
}
