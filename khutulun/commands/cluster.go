package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(clusterCommand)
}

var clusterCommand = &cobra.Command{
	Use:   "cluster",
	Short: "Work with clusters",
}
