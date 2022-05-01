package host

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Host) Schedule() {
	reconcile := NewReconcile()

	if identifiers, err := self.ListPackages("", "clout"); err == nil {
		for _, identifier := range identifiers {
			reconcile_ := self.ScheduleService(identifier.Namespace, identifier.Name)
			reconcile.Merge(reconcile_)
		}
	} else {
		scheduleLog.Errorf("%s", err.Error())
	}

	self.HandleReconcile(reconcile)
}

func (self *Host) ScheduleService(namespace string, serviceName string) Reconcile {
	if namespace == "" {
		namespace = "_"
	}

	scheduleLog.Infof("scheduling service: %s, %s", namespace, serviceName)

	if lock, clout, err := self.OpenClout(namespace, serviceName); err == nil {
		if coerced, err := clout.Copy(); err == nil {
			if err := self.CoerceClout(coerced); err == nil {
				reconcile1, changed1 := self.scheduleRunnables(namespace, serviceName, clout, coerced)
				reconcile2, changed2 := self.scheduleConnections(namespace, serviceName, clout, coerced)
				if changed1 || changed2 {
					if err := self.SaveClout(namespace, serviceName, clout); err != nil {
						scheduleLog.Errorf("%s", err.Error())
					}
				}
				logging.CallAndLogError(lock.Unlock, "unlock", scheduleLog)
				reconcile1.Merge(reconcile2)
				return reconcile1
			} else {
				scheduleLog.Errorf("%s", err.Error())
			}
		} else {
			scheduleLog.Errorf("%s", err.Error())
		}
	} else {
		scheduleLog.Errorf("%s", err.Error())
	}

	return NewReconcile()
}

func (self *Host) scheduleRunnables(namespace string, serviceName string, clout *cloutpkg.Clout, coerced *cloutpkg.Clout) (Reconcile, bool) {
	containers := GetCloutContainers(coerced)
	if len(containers) == 0 {
		return nil, false
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
				scheduleLog.Errorf("%s", err)
			}
		}

		scheduleLog.Infof("scheduling %s, %s, %s to %s", namespace, serviceName, container.Name, container.Host)
		reconcile.Add(container.Host, &identifier)
	}

	return reconcile, changed
}

func (self *Host) scheduleConnections(namespace string, serviceName string, clout *cloutpkg.Clout, coerced *cloutpkg.Clout) (Reconcile, bool) {
	reconcile := NewReconcile()
	return reconcile, false
}
