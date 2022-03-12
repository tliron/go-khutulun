package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(runnableCommand)
	runnableCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
	runnableCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace")
}

var runnableCommand = &cobra.Command{
	Use:   "runnable",
	Short: "Work with runnables",
}
