package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	templateCommand.AddCommand(templateDelistCommand)
}

var templateDelistCommand = &cobra.Command{
	Use:   "delist [TEMPLATE NAME]",
	Short: "Delist a template",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		delist(namespace, "template", args)
	},
}
