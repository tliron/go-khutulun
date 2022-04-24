package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	pluginCommand.AddCommand(pluginListCommand)
}

var pluginListCommand = &cobra.Command{
	Use:   "list",
	Short: "List registered plugins",
	Run: func(cmd *cobra.Command, args []string) {
		listPackages(namespace, "plugin")
	},
}
