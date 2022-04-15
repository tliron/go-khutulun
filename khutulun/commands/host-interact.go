package commands

import (
	"os"

	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/exec"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
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

		client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		var terminal_ *exec.Terminal
		if pseudoTerminal {
			terminal_, err = exec.NewTerminal()
			util.FailOnError(err)
			util.OnExitError(terminal_.Close)
		}

		identifier := []string{"host", hostName}
		environment := map[string]string{"TERM": os.Getenv("TERM")}
		err = client.Interact(identifier, os.Stdin, terminal.Stdout, terminal.Stderr, terminal_, environment, command...)
		util.FailOnError(err)
	},
}
