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
	runnableCommand.AddCommand(runnableInteractCommand)
	runnableInteractCommand.Flags().BoolVarP(&pseudoTerminal, "terminal", "t", false, "whether to create a pseudo-terminal")
}

var runnableInteractCommand = &cobra.Command{
	Use:   "interact [SERVICE NAME] [RUNNABLE NAME] -- ...",
	Short: "Interact with a runnable",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		resourceName := args[1]

		var command []string
		if len(args) > 2 {
			command = args[2:]
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

		identifier := []string{"runnable", namespace, serviceName, resourceName}
		environment := map[string]string{"TERM": os.Getenv("TERM")}
		err = client.Interact(identifier, os.Stdin, terminal.Stdout, terminal.Stderr, terminal_, environment, command...)
		util.FailOnError(err)
	},
}
