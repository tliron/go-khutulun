package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(hostCommand)
	hostCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
}

var hostCommand = &cobra.Command{
	Use:   "host",
	Short: "Work with cluster hosts",
}
