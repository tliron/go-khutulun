package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/util"
)

func init() {
	hostCommand.AddCommand(hostAddCommand)
}

var hostAddCommand = &cobra.Command{
	Use:   "add [NAME] [ADDRESS]",
	Short: "Add a host to a cluster",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		address := args[1]

		client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		err = client.AddHost(name, address)
		util.FailOnError(err)
	},
}
