package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(delegateCommand)
	delegateCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
	delegateCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace")
}

var delegateCommand = &cobra.Command{
	Use:   "delegate",
	Short: "Work with delegates",
}

func getPluginArgs(args []string) []string {
	args_ := []string{args[0]}
	if len(args) > 1 {
		args_ = append(args_, args[1:]...)
	}
	return args_
}
