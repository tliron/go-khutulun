package main

import (
	"github.com/tliron/go-khutulun/delegate"
	"github.com/tliron/go-khutulun/sdk"
	cloutpkg "github.com/tliron/go-puccini/clout"
)

func (self *Delegate) ListResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout) ([]delegate.Resource, error) {
	containers, err := sdk.GetCloutOCIContainers(coercedClout)
	if err != nil {
		return nil, err
	}
	connections, err := sdk.GetCloutConnections(coercedClout)
	if err != nil {
		return nil, err
	}
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
