package commands

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tliron/go-khutulun/configuration"
	"github.com/tliron/go-kutil/util"
)

var setDefault bool

func init() {
	clusterCommand.AddCommand(clusterSetCommand)
	clusterSetCommand.Flags().BoolVarP(&setDefault, "set-default", "d", false, "set as default cluster")
}

var clusterSetCommand = &cobra.Command{
	Use:   "set [CLUSTER NAME] [IP] [PORT]",
	Short: "Add a cluster to or update it in the list of known clusters",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		ip := args[1]
		port, err := strconv.ParseUint(args[2], 10, 0)
		util.FailOnError(err)

		client, err := configuration.LoadOrNewClient(configurationPath)
		util.FailOnError(err)
		client.Clusters[clusterName] = configuration.Cluster{
			IP:   ip,
			Port: int(port),
		}
		if setDefault {
			client.DefaultCluster = clusterName
		}

		err = client.Validate()
		util.FailOnError(err)
		err = client.Save("")
		util.FailOnError(err)
	},
}
