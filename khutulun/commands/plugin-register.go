package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	pluginCommand.AddCommand(pluginRegisterCommand)
	pluginRegisterCommand.Flags().StringVarP(&unpack, "unpack", "u", "auto", "unpack archive (\"auto\" or \"false\")")
}

var pluginRegisterCommand = &cobra.Command{
	Use:   "register [PLUGIN NAME] [[CONTENT PATH or URL]]",
	Short: "Register a plugin",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		registerPackage(namespace, "plugin", getPluginArgs(args))
	},
}
