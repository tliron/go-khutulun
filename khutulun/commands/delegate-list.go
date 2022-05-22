package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	delegateCommand.AddCommand(delegateListCommand)
}

var delegateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List registered delegates",
	Run: func(cmd *cobra.Command, args []string) {
		listPackages(namespace, "delegate")
	},
}
