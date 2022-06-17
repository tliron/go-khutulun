package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) ListResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout) ([]delegate.Resource, error) {
	processes := sdk.GetCloutProcesses(coercedClout)
	var resources []delegate.Resource

	for _, container := range processes {
		resources = append(resources, delegate.Resource{
			Type: "activity",
			Name: container.Name,
			Host: container.Host,
		})
	}

	return resources, nil
}
