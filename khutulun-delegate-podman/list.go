package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) ListResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout) ([]delegate.Resource, error) {
	containers := sdk.GetCloutContainers(coercedClout)
	connections := sdk.GetCloutConnections(coercedClout)
	var resources []delegate.Resource

	for _, container := range containers {
		resources = append(resources, delegate.Resource{
			Type: "activity",
			Name: container.Name,
			Host: container.Host,
		})
	}

	for _, connection := range connections {
		resources = append(resources, delegate.Resource{
			Type: "connection",
			Name: connection.Name,
		})
	}

	return resources, nil
}
