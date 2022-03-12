package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/khutulun/configuration"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

func init() {
	clusterCommand.AddCommand(clusterDefaultCommand)
}

var clusterDefaultCommand = &cobra.Command{
	Use:   "default [[CLUSTER NAME]]",
	Short: "Get or set the default cluster",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := configuration.LoadOrNewClient(configurationPath)
		util.FailOnError(err)

		if len(args) == 0 {
			if client.DefaultCluster != "" {
				terminal.Println(client.DefaultCluster)
			}
		} else {
			client.DefaultCluster = args[0]

			err = client.Validate()
			util.FailOnError(err)
			err = client.Save("")
			util.FailOnError(err)
		}
	},
}
