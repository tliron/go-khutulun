package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(templateCommand)
	templateCommand.PersistentFlags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
	templateCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace")
}

var templateCommand = &cobra.Command{
	Use:   "template",
	Short: "Work with templates",
}
