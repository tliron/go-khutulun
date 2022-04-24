package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	profileCommand.AddCommand(profileListCommand)
}

var profileListCommand = &cobra.Command{
	Use:   "list",
	Short: "List registered profiles",
	Run: func(cmd *cobra.Command, args []string) {
		listPackages(namespace, "profile")
	},
}
