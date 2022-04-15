package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/util"
)

func init() {
	serviceCommand.AddCommand(serviceDeleteCommand)
}

var serviceDeleteCommand = &cobra.Command{
	Use:   "delete [SERVICE NAME]",
	Short: "Delete a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		err = client.RemoveArtifact(namespace, "clout", args[0])
		util.FailOnError(err)
	},
}
