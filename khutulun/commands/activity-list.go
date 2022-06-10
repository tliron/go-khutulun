package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	activityCommand.AddCommand(activityListCommand)
}

var activityListCommand = &cobra.Command{
	Use:   "list [[SERVICE NAME]]",
	Short: "List activities",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listResources("activity", args)
	},
}
