package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/go-khutulun/configuration"
	"github.com/tliron/go-kutil/util"
)

func init() {
	clusterCommand.AddCommand(clusterListCommand)
}

var clusterListCommand = &cobra.Command{
	Use:   "list",
	Short: "List known clusters",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := configuration.LoadOrNewClient(configurationPath)
		util.FailOnError(err)
		err = Transcriber().Write(client.Clusters)
		util.FailOnError(err)
	},
}
