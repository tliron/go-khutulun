package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/go-khutulun/client"
	"github.com/tliron/go-kutil/util"
)

func init() {
	hostCommand.AddCommand(hostListCommand)
}

var hostListCommand = &cobra.Command{
	Use:   "list",
	Short: "List hosts in a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		hosts, err := client.ListHosts()
		util.FailOnError(err)
		if len(hosts) > 0 {
			err = Transcriber().Write(hosts)
			util.FailOnError(err)
		}
	},
}
