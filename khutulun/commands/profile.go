package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(profileCommand)
	profileCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
	profileCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace")
}

var profileCommand = &cobra.Command{
	Use:   "profile",
	Short: "Work with profiles",
}
