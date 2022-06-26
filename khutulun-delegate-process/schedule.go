package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) Schedule(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	processes := sdk.GetCloutProcesses(coercedClout)
	if len(processes) == 0 {
		return nil, nil, nil
	}

	var next []delegate.Next

	var changed bool
	for _, process := range processes {
		if process.Host == "" {
			// Schedule (TODO)
			process.Host = self.host

			if _, capability, err := process.Find(clout); err == nil {
				if sdk.Schedule(capability, process.Host) {
					changed = true
				}
			} else {
				log.Errorf("%s", err)
			}
		}

		log.Infof("scheduling %s/%s->%s to host %s", namespace, serviceName, process.Name, process.Host)
		next = delegate.AppendNext(next, process.Host, "reconcile", namespace, serviceName)
	}

	if changed {
		return clout, next, nil
	} else {
		return nil, next, nil
	}
}
