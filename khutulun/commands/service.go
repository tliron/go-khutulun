package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(serviceCommand)
	serviceCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
	serviceCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace")
}

var serviceCommand = &cobra.Command{
	Use:   "service",
	Short: "Work with services",
}
