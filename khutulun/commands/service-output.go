package commands

import (
	"strings"

	"github.com/spf13/cobra"
	clientpkg "github.com/tliron/go-khutulun/client"
	"github.com/tliron/go-kutil/util"
	cloutpkg "github.com/tliron/go-puccini/clout"
	cloututil "github.com/tliron/go-puccini/clout/util"
)

func init() {
	serviceCommand.AddCommand(serviceOutputCommand)
}

var serviceOutputCommand = &cobra.Command{
	Use:   "output [SERVICE NAME] [[OUTPUT NAME]]",
	Short: "Get a service's Clout output or outputs",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			serviceOutput(args[0], &args[1])
		} else {
			serviceOutput(args[0], nil)
		}
	},
}

func serviceOutput(serviceName string, outputName *string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	var buffer strings.Builder
	err = client.GetPackageFile(namespace, "service", serviceName, "clout.yaml", true, &buffer)
	util.FailOnError(err)

	var clout *cloutpkg.Clout
	clout, err = cloutpkg.Read(strings.NewReader(buffer.String()), "yaml")
	util.FailOnError(err)

	if outputs, ok := cloututil.GetToscaOutputs(clout.Properties); ok {
		if outputName != nil {
			if output, ok := outputs[*outputName]; ok {
				Transcriber().Write(output)
			}
		} else {
			Transcriber().Write(outputs)
		}
	}
}
