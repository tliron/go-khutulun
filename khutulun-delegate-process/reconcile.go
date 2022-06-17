package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

// systemctl --machine user@.host --user
// https://superuser.com/a/1461905

func (self *Delegate) Reconcile(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	processes := sdk.GetCloutProcesses(coercedClout)
	if len(processes) == 0 {
		return nil, nil, nil
	}

	for _, process := range processes {
		if process.Host == self.host {
			//format.WriteGo(container, logging.GetWriter(), " ")
			if err := self.CreateProcessUserService(process); err != nil {
				log.Errorf("instantiate: %s", err.Error())
			}
		}
	}

	return nil, nil, nil
}

func (self *Delegate) CreateProcessUserService(container *sdk.Process) error {
	//serviceName := fmt.Sprintf("%s-%s.service", servicePrefix, container.Name)
	return nil
}
