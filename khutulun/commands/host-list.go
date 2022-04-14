package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	formatpkg "github.com/tliron/kutil/format"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func init() {
	hostCommand.AddCommand(hostListCommand)
}

var hostListCommand = &cobra.Command{
	Use:   "list",
	Short: "List hosts in a cluster",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := clientpkg.NewClient(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		hosts, err := client.ListHosts()
		util.FailOnError(err)
		if len(hosts) > 0 {
			err = formatpkg.Print(hosts, format, terminal.Stdout, strict, pretty)
			util.FailOnError(err)
		}
	},
}
