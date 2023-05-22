package commands

import (
	"github.com/spf13/cobra"
	"github.com/tliron/exturl"
	"github.com/tliron/go-ard"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/util"
	"github.com/tliron/yamlkeys"
)

var template string
var inputs map[string]string
var inputsUrl string
var async bool

var inputValues = make(map[string]any)

func init() {
	serviceCommand.AddCommand(serviceDeployCommand)
	serviceDeployCommand.Flags().StringVarP(&template, "template", "t", "", "registered template name (defaults to same name as service)")
	serviceDeployCommand.Flags().StringToStringVarP(&inputs, "input", "i", nil, "specify an input (format is name=YAML)")
	serviceDeployCommand.Flags().StringVarP(&inputsUrl, "inputs", "m", "", "load inputs from a PATH or URL to YAML content")
	serviceDeployCommand.Flags().BoolVarP(&async, "async", "a", false, "if true will not wait for deployment to finish")
}

var serviceDeployCommand = &cobra.Command{
	Use:   "deploy [SERVICE NAME]",
	Short: "Deploy a service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ParseInputs()

		serviceName := args[0]
		if template == "" {
			template = serviceName
		}

		client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
		util.FailOnError(err)
		util.OnExitError(client.Close)

		if !async {
			client.Timeout *= 10
		}

		err = client.DeployService(namespace, serviceName, namespace, template, inputValues, async)
		util.FailOnError(err)
	},
}

func ParseInputs() {
	if inputsUrl != "" {
		log.Infof("load inputs from %q", inputsUrl)

		urlContext := exturl.NewContext()
		util.OnExitError(urlContext.Release)

		url, err := exturl.NewValidURL(inputsUrl, nil, urlContext)
		util.FailOnError(err)
		reader, err := url.Open()
		util.FailOnError(err)
		defer reader.Close()
		data, err := yamlkeys.DecodeAll(reader)
		util.FailOnError(err)
		for _, data_ := range data {
			if map_, ok := data_.(ard.Map); ok {
				for key, value := range map_ {
					inputValues[yamlkeys.KeyString(key)] = value
				}
			} else {
				util.Failf("malformed inputs in %q", inputsUrl)
			}
		}
	}

	if inputs != nil {
		for name, input := range inputs {
			input_, _, err := ard.DecodeYAML(input, false)
			util.FailOnError(err)
			inputValues[name] = input_
		}
	}
}
