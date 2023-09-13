package commands

import (
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/kutil/util"
)

func listResources(type_ string, args []string) {
	client, err := clientpkg.NewClientFromConfiguration(configurationPath, clusterName)
	util.FailOnError(err)
	util.OnExitError(client.Close)

	var service string
	if len(args) > 0 {
		service = args[0]
	}

	resources, err := client.ListResources(namespace, service, type_)
	util.FailOnError(err)
	if len(resources) > 0 {
		err = Transcriber().Write(resources)
		util.FailOnError(err)
	}
}
