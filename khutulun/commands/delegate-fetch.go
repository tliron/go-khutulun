package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	delegateCommand.AddCommand(delegateFetchCommand)
}

var delegateFetchCommand = &cobra.Command{
	Use:   "fetch [DELEGATE NAME]",
	Short: "Fetch a delegate's content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fetchPackage(namespace, "delegate", getPluginArgs(args))
	},
}
