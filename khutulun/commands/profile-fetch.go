package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	profileCommand.AddCommand(profileFetchCommand)
}

var profileFetchCommand = &cobra.Command{
	Use:   "fetch [PROFILE NAME] [[PATH]]",
	Short: "List or fetch a profile's content",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		fetchBundle(namespace, "profile", args)
	},
}
