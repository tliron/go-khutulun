package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	activityCommand.AddCommand(activityInteractCommand)
	activityInteractCommand.Flags().BoolVarP(&pseudoTerminal, "terminal", "t", false, "whether to create a pseudo-terminal")
	activityInteractCommand.Flags().BoolVarP(&forwardExitCode, "forward-exit", "e", true, "whether to forward the remote exit code")
}

var activityInteractCommand = &cobra.Command{
	Use:   "interact [SERVICE NAME] [ACTIVITY NAME] -- ...",
	Short: "Interact with an activity",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		resourceName := args[1]

		var command []string
		if len(args) > 2 {
			command = args[2:]
		}

		interact([]string{"activity", namespace, serviceName, resourceName}, command)
	},
}
