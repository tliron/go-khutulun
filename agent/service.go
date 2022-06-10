package agent

import (
	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Agent) DeployService(templateNamespace string, templateName string, serviceNamespace string, serviceName string, async bool) error {
	if _, problems, err := self.CompileTOSCA(templateNamespace, templateName, serviceNamespace, serviceName); err == nil {
		schedule := func() {
			delegates := self.NewDelegates()
			defer delegates.Release()

			self.ProcessService(serviceNamespace, serviceName, "schedule", delegates)
		}

		if async {
			go schedule()
		} else {
			schedule()
		}

		return nil
	} else {
		if problems != nil {
			return problems.WithError(nil, false)
		} else {
			return err
		}
	}
}

func (self *Agent) ProcessAllServices(phase string) {
	if identifiers, err := self.ListPackages("", "service"); err == nil {
		delegates := self.NewDelegates()
		defer delegates.Release()

		for _, identifier := range identifiers {
			self.ProcessService(identifier.Namespace, identifier.Name, phase, delegates)
		}
	} else {
		delegateLog.Errorf("%s", err.Error())
	}
}

func (self *Agent) ProcessService(namespace string, serviceName string, phase string, delegates *Delegates) {
	if namespace == "" {
		namespace = "_"
	}

	delegateLog.Infof("processing service %s: %s/%s", phase, namespace, serviceName)

	var next []delegatepkg.Next

	if lock, clout, err := self.OpenServiceClout(namespace, serviceName); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", delegateLog)

		if coercedClout, err := self.CoerceClout(clout, true); err == nil {

			var saveClout *cloutpkg.Clout

			if phase == "schedule" { // TODO: move to its own phase?
				if self.Instantiate(clout, coercedClout) {
					// Re-coerce
					if coercedClout, err = self.CoerceClout(clout, true); err != nil {
						delegateLog.Errorf("%s", err.Error())
						return
					}
					saveClout = clout
				}
			}

			delegates.Fill(namespace, coercedClout)

			for _, delegate := range delegates.All() {
				//for _, delegate_ := range delegates.delegates {
				//delegate := delegate_.Delegate()
				//log.Noticef("!!!!!!!!!!!!! delegate: %s", delegate_.Name())
				if changedClout, next_, err := delegate.ProcessService(namespace, serviceName, phase, clout, coercedClout); err == nil {
					next = delegatepkg.MergeNexts(next, next_)

					if changedClout != nil {
						clout = changedClout
						saveClout = changedClout

						if coercedClout, err = self.CoerceClout(clout, true); err != nil {
							delegateLog.Errorf("%s", err.Error())
						}
					}
				} else {
					delegateLog.Errorf("%s", err.Error())
				}
			}

			if saveClout != nil {
				if err := self.SaveServiceClout(namespace, serviceName, saveClout); err != nil {
					delegateLog.Errorf("%s", err.Error())
				}
			}
		} else {
			delegateLog.Errorf("%s", err.Error())
		}
	} else {
		delegateLog.Errorf("%s", err.Error())
	}

	self.ProcessNext(next, delegates)
}

func (self *Agent) ProcessNext(next []delegatepkg.Next, delegates *Delegates) {
	//log.Infof("NEXT: %v", next)
	for _, next_ := range next {
		if next_.Host == self.host {
			self.ProcessService(next_.Namespace, next_.ServiceName, next_.Phase, delegates)
		} else {
			if err := self.gossip.SendJSON(next_.Host, NewProcessServiceCommand(next_.Namespace, next_.ServiceName, next_.Phase)); err != nil {
				delegateLog.Errorf("%s", err.Error())
			}
		}
	}
}
