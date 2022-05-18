package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/khutulun/dashboard"
)

func init() {
	rootCommand.AddCommand(dashboardCommand)
}

var dashboardCommand = &cobra.Command{
	Use:   "dashboard",
	Short: "Dashboard TUI",
	Run: func(cmd *cobra.Command, args []string) {
		dashboard.Dashboard()
	},
}
