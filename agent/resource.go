package agent

import (
	"os"

	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

type ResourceIdentifier struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Service   string `json:"service" yaml:"service"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
	Host      string `json:"host" yaml:"host"`
}

func (self *Agent) ListResources(namespace string, serviceName string, type_ string) ([]ResourceIdentifier, error) {
	var resources []ResourceIdentifier

	var packages []PackageIdentifier
	if serviceName == "" {
		var err error
		if packages, err = self.ListPackages(namespace, "clout"); err != nil {
			return nil, err
		}
	} else {
		packages = []PackageIdentifier{
			{
				Namespace: namespace,
				Type:      "clout",
				Name:      serviceName,
			},
		}
	}

	for _, package_ := range packages {
		if lock, clout, err := self.OpenClout(package_.Namespace, package_.Name); err == nil {
			logging.CallAndLogError(lock.Unlock, "unlock", log)
			if err := self.CoerceClout(clout); err == nil {
				if resources_, err := self.getResources(package_.Namespace, package_.Name, clout, type_); err == nil {
					for _, resource := range resources_ {
						if resource.Type == type_ {
							resources = append(resources, ResourceIdentifier{
								Namespace: package_.Namespace,
								Service:   package_.Name,
								Type:      type_,
								Name:      resource.Name,
								Host:      resource.Host,
							})
						}
					}
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (self *Agent) getResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout, type_ string) ([]delegatepkg.Resource, error) {
	delegates := self.NewDelegates()
	delegates.Fill(namespace, coercedClout)
	defer delegates.Release()

	var resources []delegatepkg.Resource

	for _, delegate := range delegates.All() {
		if resources_, err := delegate.ListResources(namespace, serviceName, coercedClout); err == nil {
			resources = append(resources, resources_...)
		} else {
			return nil, err
		}
	}

	return resources, nil
}
