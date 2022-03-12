package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	profileCommand.AddCommand(profileRegisterCommand)
}

var profileRegisterCommand = &cobra.Command{
	Use:   "register [PROFILE NAME] [[CONTENT PATH or URL]]",
	Short: "Register a profile",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		registerArtifact(namespace, "profile", args)
	},
}
