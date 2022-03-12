package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	templateCommand.AddCommand(templateListCommand)
}

var templateListCommand = &cobra.Command{
	Use:   "list",
	Short: "List registered templates",
	Run: func(cmd *cobra.Command, args []string) {
		listArtifacts(namespace, "template")
	},
}
