package commands

import (
	"os"

	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
	"golang.org/x/term"
)

func init() {
	hostCommand.AddCommand(hostInteractCommand)
	hostInteractCommand.Flags().BoolVarP(&pseudoTerminal, "terminal", "t", false, "whether to create a pseudo-terminal")
}

var hostInteractCommand = &cobra.Command{
	Use:   "interact [HOST NAME] -- ...",
	Short: "Interact with a host in a cluster",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hostName := args[0]

		var command []string
		if len(args) > 1 {
			command = args[1:]
		}

		client, err := clientpkg.NewClient(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		if pseudoTerminal {
			termState, err := term.MakeRaw(int(os.Stdin.Fd()))
			util.FailOnError(err)
			util.OnExitError(func() error {
				return term.Restore(int(os.Stdin.Fd()), termState)
			})
		}

		identifier := []string{"host", hostName}
		err = client.Interact(identifier, os.Stdin, terminal.Stdout, terminal.Stderr, pseudoTerminal, command...)
		util.FailOnError(err)
	},
}
