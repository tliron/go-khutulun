package conductor

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Conductor) Schedule() {
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

func (self *Conductor) ScheduleService(namespace string, serviceName string) Reconcile {
	if namespace == "" {
		namespace = "_"
	}

	scheduleLog.Infof("scheduling service: %s, %s", namespace, serviceName)

	if lock, clout, err := self.OpenClout(namespace, serviceName); err == nil {
		reconcile, changed := self.scheduleRunnables(namespace, serviceName, clout)
		if changed {
			if err := self.SaveClout(namespace, serviceName, clout); err != nil {
				scheduleLog.Errorf("%s", err.Error())
			}
		}
		logging.CallAndLogError(lock.Unlock, "unlock", scheduleLog)
		return reconcile
	} else {
		scheduleLog.Errorf("%s", err.Error())
	}

	return NewReconcile()
}

func (self *Conductor) scheduleRunnables(namespace string, serviceName string, clout *cloutpkg.Clout) (Reconcile, bool) {
	var coerced *cloutpkg.Clout
	var err error
	if coerced, err = clout.Copy(); err == nil {
		if err = self.CoerceClout(coerced); err != nil {
			scheduleLog.Errorf("%s", err.Error())
			return nil, false
		}
	} else {
		scheduleLog.Errorf("%s", err.Error())
		return nil, false
	}

	containers := self.getResources(coerced, "runnable")
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
