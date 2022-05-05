package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	runnableCommand.AddCommand(runnableInteractCommand)
	runnableInteractCommand.Flags().BoolVarP(&pseudoTerminal, "terminal", "t", false, "whether to create a pseudo-terminal")
	runnableInteractCommand.Flags().BoolVarP(&forwardExitCode, "forward-exit", "e", true, "whether to forward the remote exit code")
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

		interact([]string{"runnable", namespace, serviceName, resourceName}, command)
	},
}
