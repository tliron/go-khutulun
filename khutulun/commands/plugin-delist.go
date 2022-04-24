package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	pluginCommand.AddCommand(pluginDelistCommand)
}

var pluginDelistCommand = &cobra.Command{
	Use:   "delist [PLUGIN TYPE] [PLUGIN NAME]",
	Short: "Delist a plugin",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		delistPackage(namespace, "plugin", getPluginArgs(args))
	},
}
