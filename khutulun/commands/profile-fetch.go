package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	profileCommand.AddCommand(profileFetchCommand)
}

var profileFetchCommand = &cobra.Command{
	Use:   "fetch [PROFILE NAME]",
	Short: "Fetch a profile's content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fetchArtifact(namespace, "profile", args)
	},
}
