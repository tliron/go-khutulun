package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	pluginCommand.AddCommand(pluginDelistCommand)
}

var pluginDelistCommand = &cobra.Command{
	Use:   "delist [PLUGIN NAME]",
	Short: "Delist a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		delistPackage(namespace, "plugin", getPluginArgs(args))
	},
}
