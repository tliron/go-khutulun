package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	hostCommand.AddCommand(hostInteractCommand)
	hostInteractCommand.Flags().BoolVarP(&pseudoTerminal, "terminal", "t", false, "whether to create a pseudo-terminal")
	hostInteractCommand.Flags().BoolVarP(&forwardExitCode, "forward", "e", true, "whether to forward the remote exit code")
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

		interact([]string{"host", hostName}, command)
	},
}
