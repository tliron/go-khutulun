package commands

import (
	"os"

	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/khutulun/client"
	terminalutil "github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	runnableCommand.AddCommand(runnableInteractCommand)
}

var runnableInteractCommand = &cobra.Command{
	Use:   "interact [SERVICE NAME] [RUNNABLE NAME]",
	Short: "Interact with a runnable",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		resourceName := args[1]

		client, err := clientpkg.NewClient(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		// See: https://stackoverflow.com/a/54423725
		/*exec.Command("/usr/bin/stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
		exec.Command("/usr/bin/stty", "-F", "/dev/tty", "-echo").Run()
		util.OnExit(func() {
			exec.Command("/usr/bin/stty", "-F", "/dev/tty", "echo").Run()
		})*/

		termState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
		util.FailOnError(err)
		util.OnExit(func() {
			terminal.Restore(int(os.Stdin.Fd()), termState)
		})

		err = client.InteractRunnable(namespace, serviceName, resourceName, os.Stdin, terminalutil.Stdout, terminalutil.Stderr)
		util.FailOnError(err)
	},
}
