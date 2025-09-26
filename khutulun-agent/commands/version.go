package commands

import (
	"github.com/tliron/go-kutil/cobra"
)

func init() {
	rootCommand.AddCommand(cobra.NewVersionCommand(toolName))
}
