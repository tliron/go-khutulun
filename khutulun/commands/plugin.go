package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(pluginCommand)
	pluginCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
	pluginCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace")
}

var pluginCommand = &cobra.Command{
	Use:   "plugin",
	Short: "Work with plugins",
}

func getPluginArgs(args []string) []string {
	args_ := []string{args[0]}
	if len(args) > 1 {
		args_ = append(args_, args[1:]...)
	}
	return args_
}
