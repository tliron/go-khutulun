package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	profileCommand.AddCommand(profileDelistCommand)
}

var profileDelistCommand = &cobra.Command{
	Use:   "delist [PROFILE NAME]",
	Short: "Delist a profile",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		delist(namespace, "profile", args)
	},
}
