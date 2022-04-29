package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(connectionCommand)
	connectionCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
	connectionCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace")
}

var connectionCommand = &cobra.Command{
	Use:   "connection",
	Short: "Work with connections",
}
