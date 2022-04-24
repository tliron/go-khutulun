package conductor

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
}

func (self *Conductor) ListResources(namespace string, serviceName string, type_ string) ([]ResourceIdentifier, error) {
	var resources []ResourceIdentifier

	var identifiers []PackageIdentifier
	if serviceName == "" {
		var err error
		if identifiers, err = self.ListPackages(namespace, "clout"); err != nil {
			return nil, err
		}
	} else {
		identifiers = []PackageIdentifier{
			{
				Namespace: namespace,
				Type:      "clout",
				Name:      serviceName,
			},
		}
	}

	for _, identifier := range identifiers {
		if lock, clout, err := self.OpenClout(identifier.Namespace, identifier.Name); err == nil {
			logging.CallAndLogError(lock.Unlock, "unlock", log)

			if err := self.CoerceClout(clout); err == nil {
				for _, resource := range self.getResources(clout, type_) {
					resources = append(resources, ResourceIdentifier{
						Namespace: identifier.Namespace,
						Service:   identifier.Name,
						Type:      type_,
						Name:      resource.Name,
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

func (self *Conductor) getResources(clout *cloutpkg.Clout, type_ string) []Container {
	var containers []Container
	for _, vertex := range clout.Vertexes {
		containers = append(containers, GetContainers(vertex)...)
	}
	return containers
}
