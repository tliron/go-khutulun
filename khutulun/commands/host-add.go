package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	hostCommand.AddCommand(hostAddCommand)
}

var hostAddCommand = &cobra.Command{
	Use:   "add",
	Short: "Add a host to a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		// install conductor as a systemd service
	},
}
