package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(activityCommand)
	activityCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
	activityCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace")
}

var activityCommand = &cobra.Command{
	Use:   "activity",
	Short: "Work with activities",
}
