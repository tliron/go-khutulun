package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	memberCommand.AddCommand(memberAddCommand)
}

var memberAddCommand = &cobra.Command{
	Use:   "add",
	Short: "Add a member to a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// install conductor as a systemd service
	},
}
