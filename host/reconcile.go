package host

import (
	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Host) Reconcile() {
	if identifiers, err := self.ListPackages("", "clout"); err == nil {
		for _, identifier := range identifiers {
			self.ReconcileService(identifier.Namespace, identifier.Name)
		}
	} else {
		reconcileLog.Errorf("%s", err.Error())
	}
}

func (self *Host) ReconcileService(namespace string, serviceName string) {
	if namespace == "" {
		namespace = "_"
	}

	scheduleLog.Infof("reconciling service: %s, %s", namespace, serviceName)

	if lock, clout, err := self.OpenClout(namespace, serviceName); err == nil {
		logging.CallAndLogError(lock.Unlock, "unlock", reconcileLog)

		if err := self.CoerceClout(clout); err == nil {
			self.reconcileRunnables(clout)
			self.reconcileConnections(clout)
		} else {
			reconcileLog.Errorf("%s", err.Error())
		}
	} else {
		reconcileLog.Errorf("%s", err.Error())
	}
}

func (self *Host) reconcileRunnables(clout *cloutpkg.Clout) {
	containers := GetCloutContainers(clout)
	if len(containers) == 0 {
		return
	}

	go func() {
		var runnable plugin.Runnable
		for _, container := range containers {
			if self.host == container.Host {
				if runnable == nil {
					name := "runnable.podman"
					command := self.getPackageMainFile("common", "plugin", name)
					client := plugin.NewRunnableClient(name, command)
					defer client.Close()
					var err error
					if runnable, err = client.Runnable(); err != nil {
						reconcileLog.Errorf("plugin: %s", err.Error())
						return
					}
				}

				if err := runnable.Instantiate(container.Container); err != nil {
					reconcileLog.Errorf("instantiate: %s", err.Error())
				}
			}
		}
	}()
}

func (self *Host) reconcileConnections(clout *cloutpkg.Clout) {
}

//
// Reconcile
//

type Reconcile map[string]*ServiceIdentifiers

func NewReconcile() Reconcile {
	return make(map[string]*ServiceIdentifiers)
}

func (self Reconcile) Add(host string, identifier *ServiceIdentifier) bool {
	var identifiers *ServiceIdentifiers
	var ok bool
	if identifiers, ok = self[host]; !ok {
		identifiers = NewServiceIdentifiers()
		self[host] = identifiers
	}
	return identifiers.Add(identifier)
}

func (self Reconcile) Merge(reconcile Reconcile) bool {
	var added bool
	for host, identifiers := range reconcile {
		for _, identifier := range identifiers.List {
			if self.Add(host, identifier) {
				added = true
			}
		}
	}
	return added
}

func (self *Host) HandleReconcile(reconcile Reconcile) {
	for host, identifiers := range reconcile {
		if self.host == host {
			for _, identifier := range identifiers.List {
				self.ReconcileService(identifier.Namespace, identifier.Name)
			}
		} else if self.gossip != nil {
			reconcileLog.Infof("sending reconcile command to: %s", host)
			command := make(map[string]any)
			command["command"] = RECONCILE_SERVICES
			command["identifiers"] = identifiers.List
			if _, err := self.gossip.SendJSON(host, command); err != nil {
				reconcileLog.Errorf("%s", err.Error())
			}
		}
	}
}
