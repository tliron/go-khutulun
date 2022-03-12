package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func init() {
	clusterCommand.AddCommand(clusterVersionCommand)
	clusterVersionCommand.Flags().StringVarP(&clusterName, "cluster", "c", "", "cluster to access")
}

var clusterVersionCommand = &cobra.Command{
	Use:   "version",
	Short: "Get a cluster's version",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := clientpkg.NewClient(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		version, err := client.GetVersion()
		util.FailOnError(err)

		terminal.Println(version)
	},
}
