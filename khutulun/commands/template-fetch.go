package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	templateCommand.AddCommand(templateFetchCommand)
}

var templateFetchCommand = &cobra.Command{
	Use:   "fetch [TEMPLATE NAME] [[PATH]]",
	Short: "List or fetch a template's content",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		fetchPackage(namespace, "template", args)
	},
}
