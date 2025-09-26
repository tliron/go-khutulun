package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/go-khutulun/client"
	"github.com/tliron/go-khutulun/dashboard"
	"github.com/tliron/go-kutil/util"
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
