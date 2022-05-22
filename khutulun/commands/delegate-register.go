package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	delegateCommand.AddCommand(delegateRegisterCommand)
	delegateRegisterCommand.Flags().StringVarP(&unpack, "unpack", "u", "auto", "unpack archive (\"auto\" or \"false\")")
}

var delegateRegisterCommand = &cobra.Command{
	Use:   "register [DELEGATE NAME] [[CONTENT PATH or URL]]",
	Short: "Register a delegate",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		registerPackage(namespace, "delegate", getPluginArgs(args))
	},
}
