package agent

import (
	contextpkg "context"

	"github.com/tliron/commonlog"
	delegatepkg "github.com/tliron/go-khutulun/delegate"
	cloutpkg "github.com/tliron/go-puccini/clout"
)

func (self *Agent) DeployService(context contextpkg.Context, templateNamespace string, templateName string, serviceNamespace string, serviceName string, async bool) error {
	if _, problems, err := self.CompileTOSCA(context, templateNamespace, templateName, serviceNamespace, serviceName); err == nil {
		schedule := func() {
			delegates := self.NewDelegates()
			defer delegates.Release()

			self.ProcessService(context, serviceNamespace, serviceName, "schedule", delegates)
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

func (self *Agent) ProcessAllServices(context contextpkg.Context, phase string) {
	if identifiers, err := self.state.ListPackages("", "service"); err == nil {
		delegates := self.NewDelegates()
		defer delegates.Release()

		for _, identifier := range identifiers {
			self.ProcessService(context, identifier.Namespace, identifier.Name, phase, delegates)
		}
	} else {
		delegateLog.Error(err.Error())
	}
}

func (self *Agent) ProcessService(context contextpkg.Context, namespace string, serviceName string, phase string, delegates *Delegates) {
	if namespace == "" {
		namespace = "_"
	}

	delegateLog.Infof("processing service %s: %s/%s", phase, namespace, serviceName)

	var next []delegatepkg.Next

	if lock, clout, err := self.state.OpenServiceClout(context, namespace, serviceName, self.urlContext); err == nil {
		defer commonlog.CallAndLogError(lock.Unlock, "unlock", delegateLog)

		if coercedClout, err := self.CoerceClout(clout, true); err == nil {

			var saveClout *cloutpkg.Clout

			if phase == "schedule" { // TODO: move to its own phase?
				if self.Instantiate(clout, coercedClout) {
					// Re-coerce
					if coercedClout, err = self.CoerceClout(clout, true); err != nil {
						delegateLog.Error(err.Error())
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
							delegateLog.Error(err.Error())
						}
					}
				} else {
					delegateLog.Error(err.Error())
				}
			}

			if saveClout != nil {
				if err := self.state.SaveServiceClout(namespace, serviceName, saveClout); err != nil {
					delegateLog.Error(err.Error())
				}
			}
		} else {
			delegateLog.Error(err.Error())
		}
	} else {
		delegateLog.Error(err.Error())
	}

	self.ProcessNext(context, next, delegates)
}

func (self *Agent) ProcessNext(context contextpkg.Context, next []delegatepkg.Next, delegates *Delegates) {
	//log.Infof("NEXT: %v", next)
	for _, next_ := range next {
		if next_.Host == self.host {
			self.ProcessService(context, next_.Namespace, next_.ServiceName, next_.Phase, delegates)
		} else {
			if err := self.gossip.SendJSON(next_.Host, NewProcessServiceCommand(next_.Namespace, next_.ServiceName, next_.Phase)); err != nil {
				delegateLog.Error(err.Error())
			}
		}
	}
}
