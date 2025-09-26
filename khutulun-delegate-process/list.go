package main

import (
	"github.com/tliron/go-khutulun/delegate"
	"github.com/tliron/go-khutulun/sdk"
	cloutpkg "github.com/tliron/go-puccini/clout"
)

func (self *Delegate) ListResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout) ([]delegate.Resource, error) {
	processes, err := sdk.GetCloutProcesses(coercedClout)
	if err != nil {
		return nil, err
	}
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
