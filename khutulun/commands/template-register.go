package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	templateCommand.AddCommand(templateRegisterCommand)
}

var templateRegisterCommand = &cobra.Command{
	Use:   "register [TEMPLATE NAME] [[CONTENT PATH or URL]]",
	Short: "Register a template",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		registerBundle(namespace, "template", args)
	},
}
