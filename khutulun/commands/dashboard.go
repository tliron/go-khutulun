package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/khutulun/dashboard"
	"github.com/tliron/kutil/util"
)

func init() {
	rootCommand.AddCommand(dashboardCommand)
}

var dashboardCommand = &cobra.Command{
	Use:   "dashboard",
	Short: "Dashboard TUI",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)
		err = dashboard.Dashboard(client)
		util.FailOnError(err)
	},
}
