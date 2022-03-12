package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(memberCommand)
	memberCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
}

var memberCommand = &cobra.Command{
	Use:   "member",
	Short: "Work with cluster members",
}
