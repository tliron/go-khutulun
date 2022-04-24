package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCommand.AddCommand(serviceFetchCommand)
}

var serviceFetchCommand = &cobra.Command{
	Use:   "fetch [SERVICE NAME]",
	Short: "Fetch a service's Clout content",
	Run: func(cmd *cobra.Command, args []string) {
		args = append(args, "clout.yaml")
		fetchPackage(namespace, "clout", args)
	},
}
