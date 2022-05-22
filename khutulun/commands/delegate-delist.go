package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	delegateCommand.AddCommand(delegateDelistCommand)
}

var delegateDelistCommand = &cobra.Command{
	Use:   "delist [DELEGATE NAME]",
	Short: "Delist a delegate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		delistPackage(namespace, "delegate", getPluginArgs(args))
	},
}
