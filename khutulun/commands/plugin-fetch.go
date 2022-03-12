package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	pluginCommand.AddCommand(pluginFetchCommand)
}

var pluginFetchCommand = &cobra.Command{
	Use:   "fetch [PLUGIN TYPE] [PLUGIN NAME]",
	Short: "Fetch a plugin's content",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fetchArtifact(namespace, "plugin", getPluginArgs(args))
	},
}
