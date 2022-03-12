package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/khutulun/configuration"
	"github.com/tliron/kutil/util"
)

func init() {
	clusterCommand.AddCommand(clusterRemoveCommand)
}

var clusterRemoveCommand = &cobra.Command{
	Use:   "remove [CLUSTER NAME]",
	Short: "Remove a cluster from list of known clusters",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]
		client, err := configuration.LoadOrNewClient(configurationPath)
		util.FailOnError(err)
		if _, ok := client.Clusters[clusterName]; ok {
			delete(client.Clusters, clusterName)
			if clusterName == client.DefaultCluster {
				client.DefaultCluster = ""
			}
			err = client.Validate()
			util.FailOnError(err)
			err = client.Save("")
			util.FailOnError(err)
		} else {
			util.Failf("unknown cluster: %q", clusterName)
		}
	},
}
