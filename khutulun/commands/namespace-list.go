package commands

import (
	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/terminal"
	formatpkg "github.com/tliron/kutil/transcribe"
	"github.com/tliron/kutil/util"
)

func init() {
	namespaceCommand.AddCommand(namespaceListCommand)
}

var namespaceListCommand = &cobra.Command{
	Use:   "list",
	Short: "List namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		namespaces, err := client.ListNamespaces()
		util.FailOnError(err)
		if len(namespaces) > 0 {
			err = formatpkg.Print(namespaces, format, terminal.Stdout, strict, pretty)
			util.FailOnError(err)
		}
	},
}
