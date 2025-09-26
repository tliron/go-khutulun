package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/go-khutulun/client"
	"github.com/tliron/go-kutil/util"
)

func init() {
	hostCommand.AddCommand(hostAddCommand)
}

var hostAddCommand = &cobra.Command{
	Use:   "add [GOSSIP ADDRESS[:PORT]]",
	Short: "Add a host to a cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gossipAddress := args[0]

		client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		err = client.AddHost(gossipAddress)
		util.FailOnError(err)
	},
}
