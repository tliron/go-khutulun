package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCommand.AddCommand(serviceListCommand)
}

var serviceListCommand = &cobra.Command{
	Use:   "list",
	Short: "List services",
	Run: func(cmd *cobra.Command, args []string) {
		listPackages(namespace, "service")
	},
}
