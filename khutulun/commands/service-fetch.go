package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCommand.AddCommand(serviceFetchCommand)
}

var serviceFetchCommand = &cobra.Command{
	Use:   "fetch [SERVICE NAME]",
	Short: "Fetch a service's Clout content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fetchArtifact(namespace, "clout", args)
	},
}
