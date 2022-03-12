package commands

import (
	"github.com/spf13/cobra"
	conductorpkg "github.com/tliron/khutulun/conductor"
	"github.com/tliron/kutil/util"
)

var statePath string

func init() {
	rootCommand.AddCommand(serverCommand)
	serverCommand.Flags().StringVarP(&statePath, "state-path", "s", "/tmp/khutulun", "state path")
}

var serverCommand = &cobra.Command{
	Use:   "server",
	Short: "Run the server",
	Run: func(cmd *cobra.Command, args []string) {
		conductor := conductorpkg.NewConductor(statePath)
		util.OnExitError(conductor.Release)
		server := conductorpkg.NewServer(conductor)
		util.OnExitError(server.Stop)
		err := server.Start(true, true, true, true)
		util.FailOnError(err)
		select {}
	},
}
