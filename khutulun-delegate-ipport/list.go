package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) ListResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout) ([]delegate.Resource, error) {
	connections := sdk.GetCloutConnections(coercedClout)
	resources := make([]delegate.Resource, len(connections))
	for index, connection := range connections {
		resources[index] = delegate.Resource{
			Type: "connection",
			Name: connection.Name,
		}
	}
	return resources, nil
}
