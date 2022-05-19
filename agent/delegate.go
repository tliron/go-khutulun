package agent

import (
	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Agent) GetDelegate() (*delegatepkg.DelegatePluginClient, delegatepkg.Delegate, error) {
	name := "podman"
	command := self.getPackageMainFile("common", "plugin", name)
	client := delegatepkg.NewDelegatePluginClient(name, command)
	if delegate, err := client.Delegate(); err == nil {
		return client, delegate, nil
	} else {
		client.Close()
		return nil, nil, err
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

	var delegate delegatepkg.Delegate
	var client *delegatepkg.DelegatePluginClient
	var err error
	if client, delegate, err = self.GetDelegate(); err == nil {
		defer client.Close()
	} else {
		delegateLog.Errorf("%s", err.Error())
		return
	}

	var clout_ *cloutpkg.Clout
	var next []delegatepkg.Next

	if lock, clout, err := self.OpenClout(namespace, serviceName); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", delegateLog)

		if coerced, err := clout.Copy(); err == nil {
			if err := self.CoerceClout(coerced); err == nil {
				if clout_, next, err = delegate.ProcessService(namespace, serviceName, phase, clout, coerced); err == nil {
					if clout_ != nil {
						if err := self.SaveClout(namespace, serviceName, clout_); err != nil {
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
	} else {
		delegateLog.Errorf("%s", err.Error())
	}

	//log.Infof("NEXT: %v", next)
	for _, next_ := range next {
		if next_.Host == self.host {
			self.ProcessService(next_.Namespace, next_.ServiceName, next_.Phase)
		} else {
			command := make(map[string]any)
			command["command"] = PROCESS_SERVICE
			command["namespace"] = next_.Namespace
			command["serviceName"] = next_.ServiceName
			command["phase"] = next_.Phase
			if err := self.gossip.SendJSON(next_.Host, command); err != nil {
				delegateLog.Errorf("%s", err.Error())
			}
		}
	}
}
