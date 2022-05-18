package main

import (
	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) Schedule(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, error) {
	containers := sdk.GetCloutContainers(coercedClout)
	if len(containers) == 0 {
		return nil, nil
	}

	identifier := ServiceIdentifier{namespace, serviceName}

	reconcile := NewReconcile()

	var changed bool
	for _, container := range containers {
		if container.Host == "" {
			// Schedule (TODO)
			container.Host = self.host

			if _, capability, err := container.Find(clout); err == nil {
				host, _ := ard.NewNode(capability).Get("attributes").Get("host").StringMap()
				host["$value"] = container.Host
				changed = true
			} else {
				log.Errorf("%s", err)
			}
		}

		log.Infof("scheduling %s/%s->%s to %s", namespace, serviceName, container.Name, container.Host)
		reconcile.Add(container.Host, &identifier)
	}

	if changed {
		return clout, nil
	} else {
		return nil, nil
	}
}

func (self *Delegate) scheduleConnections(namespace string, serviceName string, clout *cloutpkg.Clout, coerced *cloutpkg.Clout) (Reconcile, bool) {
	connections := sdk.GetCloutConnections(coerced)
	if len(connections) == 0 {
		return nil, false
	}

	//format.PrintYAML(connections, os.Stdout, false, false)

	reconcile := NewReconcile()
	return reconcile, false
}
