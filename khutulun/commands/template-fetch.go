package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	templateCommand.AddCommand(templateFetchCommand)
}

var templateFetchCommand = &cobra.Command{
	Use:   "fetch [TEMPLATE NAME]",
	Short: "Fetch a template's content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fetchArtifact(namespace, "template", args)
	},
}
