package agent

import (
	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Agent) DeployService(templateNamespace string, templateName string, serviceNamespace string, serviceName string) error {
	if _, problems, err := self.CompileTOSCA(templateNamespace, templateName, serviceNamespace, serviceName); err == nil {
		self.ProcessService(serviceNamespace, serviceName, "schedule")
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
	if identifiers, err := self.ListPackages("", "clout"); err == nil {
		for _, identifier := range identifiers {
			self.ProcessService(identifier.Namespace, identifier.Name, phase)
		}
	} else {
		delegateLog.Errorf("%s", err.Error())
	}
}

func (self *Agent) ProcessService(namespace string, serviceName string, phase string) {
	if namespace == "" {
		namespace = "_"
	}

	delegateLog.Infof("processing service %s: %s/%s", phase, namespace, serviceName)

	var next []delegatepkg.Next

	if lock, clout, err := self.OpenClout(namespace, serviceName); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", delegateLog)

		if coercedClout, err := clout.Copy(); err == nil {
			if err := self.CoerceClout(coercedClout); err == nil {

				delegates := self.NewDelegates()
				delegates.Fill(namespace, coercedClout)
				defer delegates.Release()

				var saveClout *cloutpkg.Clout

				for _, delegate := range delegates.All() {
					//for _, delegate_ := range delegates.delegates {
					//delegate := delegate_.Delegate()
					//log.Noticef("!!!!!!!!!!!!! delegate: %s", delegate_.Name())
					if clout_, next_, err := delegate.ProcessService(namespace, serviceName, phase, clout, coercedClout); err == nil {
						next = delegatepkg.MergeNexts(next, next_)

						if clout_ != nil {
							clout = clout_
							saveClout = clout_

							if coercedClout, err = clout.Copy(); err == nil {
								if err := self.CoerceClout(coercedClout); err != nil {
									delegateLog.Errorf("%s", err.Error())
								}
							}
						}
					} else {
						delegateLog.Errorf("%s", err.Error())
					}
				}

				if saveClout != nil {
					if err := self.SaveClout(namespace, serviceName, saveClout); err != nil {
						delegateLog.Errorf("%s", err.Error())
					}
				}
			} else {
				delegateLog.Errorf("%s", err.Error())
			}
		} else {
			delegateLog.Errorf("%s", err.Error())
		}
	} else {
		delegateLog.Errorf("%s", err.Error())
	}

	//log.Infof("NEXT: %v", next)
	for _, next_ := range next {
		if next_.Host == self.host {
			self.ProcessService(next_.Namespace, next_.ServiceName, next_.Phase)
		} else {
			if err := self.gossip.SendJSON(next_.Host, NewProcessServiceCommand(next_.Namespace, next_.ServiceName, next_.Phase)); err != nil {
				delegateLog.Errorf("%s", err.Error())
			}
		}
	}
}
