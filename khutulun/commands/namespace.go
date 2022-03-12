package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(namespaceCommand)
}

var namespaceCommand = &cobra.Command{
	Use:   "namespace",
	Short: "Work with cluster namespaces",
}
