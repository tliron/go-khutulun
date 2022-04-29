package host

import (
	"os"

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

func (self *Host) ListResources(namespace string, serviceName string, type_ string) ([]ResourceIdentifier, error) {
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
				for _, resource := range getResources(clout, type_) {
					resources = append(resources, ResourceIdentifier{
						Namespace: package_.Namespace,
						Service:   package_.Name,
						Type:      type_,
						Name:      resource.name,
						Host:      resource.host,
					})
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

type Resource struct {
	name string
	host string
}

func getResources(clout *cloutpkg.Clout, type_ string) []Resource {
	var resources []Resource

	switch type_ {
	case "runnable":
		for _, container := range GetCloutContainers(clout) {
			resources = append(resources, Resource{container.Name, container.Host})
		}

	case "connection":
		for _, connection := range GetCloutConnections(clout) {
			resources = append(resources, Resource{connection.Name, ""})
		}
	}

	return resources
}
