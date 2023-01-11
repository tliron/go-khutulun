package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) Schedule(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	containers, err := sdk.GetCloutOCIContainers(coercedClout)
	if err != nil {
		return nil, nil, err
	}
	if len(containers) == 0 {
		return nil, nil, nil
	}

	var next []delegate.Next

	var changed bool
	for _, container := range containers {
		if container.Host == "" {
			// Schedule (TODO)
			container.Host = self.host

			if _, capability, err := container.Find(clout); err == nil {
				if sdk.ScheduleHost(capability, container.Host) {
					changed = true
				} else {
					// TODO
				}
			} else {
				log.Errorf("%s", err)
			}
		}

		log.Infof("scheduling %s/%s->%s to host %s", namespace, serviceName, container.Name, container.Host)
		next = delegate.AppendNext(next, container.Host, "reconcile", namespace, serviceName)
	}

	if changed {
		return clout, next, nil
	} else {
		return nil, next, nil
	}
}
