package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) ListResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout) ([]delegate.Resource, error) {
	containers := sdk.GetCloutContainers(coercedClout)
	connections := sdk.GetCloutConnections(coercedClout)
	containersLength := len(containers)
	resources := make([]delegate.Resource, containersLength+len(connections))

	for index, container := range containers {
		resources[index] = delegate.Resource{
			Type: "runnable",
			Name: container.Name,
			Host: container.Host,
		}
	}

	for index, connection := range connections {
		resources[containersLength+index] = delegate.Resource{
			Type: "connection",
			Name: connection.Name,
		}
	}

	return resources, nil
}
