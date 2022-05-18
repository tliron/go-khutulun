package agent

import (
	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/logging"
)

func (self *Agent) GetDelegate() (*delegatepkg.DelegatePluginClient, delegatepkg.Delegate, error) {
	name := "runnable.podman"
	command := self.getPackageMainFile("common", "plugin", name)
	client := delegatepkg.NewDelegatePluginClient(name, command)
	if delegate, err := client.Delegate(); err == nil {
		return client, delegate, nil
	} else {
		client.Close()
		return nil, nil, err
	}
}

func (self *Agent) Delegate(phase string) {
	var delegate delegatepkg.Delegate
	if identifiers, err := self.ListPackages("", "clout"); err == nil {
		for _, identifier := range identifiers {
			if delegate == nil {
				var client *delegatepkg.DelegatePluginClient
				if client, delegate, err = self.GetDelegate(); err == nil {
					defer client.Close()
				} else {
					delegateLog.Errorf("plugin: %s", err.Error())
					return
				}
			}

			self.ProcessService(identifier.Namespace, identifier.Name, delegate, phase)
		}
	} else {
		delegateLog.Errorf("%s", err.Error())
	}
}

func (self *Agent) ProcessService(namespace string, serviceName string, delegate delegatepkg.Delegate, phase string) {
	if namespace == "" {
		namespace = "_"
	}

	delegateLog.Infof("processing service %s: %s/%s", phase, namespace, serviceName)

	if lock, clout, err := self.OpenClout(namespace, serviceName); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", delegateLog)

		if coerced, err := clout.Copy(); err == nil {
			if err := self.CoerceClout(coerced); err == nil {
				if clout_, err := delegate.ProcessService(namespace, serviceName, phase, clout, coerced); err == nil {
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
}
