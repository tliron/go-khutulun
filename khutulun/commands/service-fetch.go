package commands

import (
	"github.com/spf13/cobra"
)

var coerce bool

func init() {
	serviceCommand.AddCommand(serviceFetchCommand)
	serviceFetchCommand.Flags().BoolVarP(&coerce, "coerce", "e", false, "whether to coerce the Clout")
}

var serviceFetchCommand = &cobra.Command{
	Use:   "fetch [SERVICE NAME]",
	Short: "Fetch a service's Clout content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		args = append(args, "clout.yaml")
		fetchPackage(namespace, "clout", args)
	},
}
