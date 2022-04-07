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

		client, err := clientpkg.NewClient(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		if pseudoTerminal {
			// See: https://stackoverflow.com/a/54423725
			/*exec.Command("/usr/bin/stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
			exec.Command("/usr/bin/stty", "-F", "/dev/tty", "-echo").Run()
			util.OnExit(func() {
				exec.Command("/usr/bin/stty", "-F", "/dev/tty", "echo").Run()
			})*/

			termState, err := term.MakeRaw(int(os.Stdin.Fd()))
			util.FailOnError(err)
			util.OnExitError(func() error {
				return term.Restore(int(os.Stdin.Fd()), termState)
			})
		}

		identifier := []string{"runnable", namespace, serviceName, resourceName}
		err = client.Interact(identifier, os.Stdin, terminal.Stdout, terminal.Stderr, pseudoTerminal, command...)
		util.FailOnError(err)
	},
}
