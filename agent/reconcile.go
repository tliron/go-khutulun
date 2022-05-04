package agent

import (
	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Agent) Reconcile() {
	if identifiers, err := self.ListPackages("", "clout"); err == nil {
		for _, identifier := range identifiers {
			self.ReconcileService(identifier.Namespace, identifier.Name)
		}
	} else {
		reconcileLog.Errorf("%s", err.Error())
	}
}

func (self *Agent) ReconcileService(namespace string, serviceName string) {
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

func (self *Agent) reconcileRunnables(clout *cloutpkg.Clout) {
	containers := GetCloutContainers(clout)
	if len(containers) == 0 {
		return
	}

	go func() {
		var delegate delegatepkg.Delegate
		for _, container := range containers {
			if self.host == container.Host {
				if delegate == nil {
					name := "runnable.podman"
					command := self.getPackageMainFile("common", "plugin", name)
					client := delegatepkg.NewDelegatePluginClient(name, command)
					defer client.Close()
					var err error
					if delegate, err = client.Delegate(); err != nil {
						reconcileLog.Errorf("plugin: %s", err.Error())
						return
					}
				}

				if err := delegate.Instantiate(container.Container); err != nil {
					reconcileLog.Errorf("instantiate: %s", err.Error())
				}
			}
		}
	}()
}

func (self *Agent) reconcileConnections(clout *cloutpkg.Clout) {
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

func (self *Agent) HandleReconcile(reconcile Reconcile) {
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
