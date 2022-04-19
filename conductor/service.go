package conductor

import (
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Conductor) GetServiceClout(namespace string, serviceName string) (*cloutpkg.Clout, error) {
	if lock, err := self.lockBundle(namespace, "clout", serviceName, false); err == nil {
		defer func() {
			if err := lock.Unlock(); err != nil {
				grpcLog.Errorf("unlock: %s", err.Error())
			}
		}()

		cloutPath := self.getBundleMainFile(namespace, "clout", serviceName)
		return cloutpkg.Load(cloutPath, "yaml")
	} else {
		return nil, err
	}
}

func (self *Conductor) DeployService(templateNamespace string, templateName string, serviceNamespace string, serviceName string) error {
	if _, problems, err := self.CompileTosca(templateNamespace, templateName, serviceNamespace, serviceName); err == nil {
		self.Reconcile()
		return nil
	} else {
		if problems != nil {
			return problems.WithError(nil, false)
		} else {
			return err
		}
	}
}
