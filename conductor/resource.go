package conductor

import (
	"os"

	"github.com/tliron/khutulun/plugin"
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

	var artifacts []ArtifactIdentifier
	if serviceName == "" {
		var err error
		if artifacts, err = self.ListArtifacts(namespace, "clout"); err != nil {
			return nil, err
		}
	} else {
		artifacts = []ArtifactIdentifier{
			{
				Namespace: namespace,
				Type:      "clout",
				Name:      serviceName,
			},
		}
	}

	for _, artifact := range artifacts {
		if clout, err := self.GetClout(artifact.Namespace, artifact.Name, true); err == nil {
			for _, resource := range self.getResources(clout, type_) {
				resources = append(resources, ResourceIdentifier{
					Namespace: artifact.Namespace,
					Service:   artifact.Name,
					Type:      type_,
					Name:      resource.Name,
				})
			}
		} else {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	return resources, nil
}

func (self *Conductor) getResources(clout *cloutpkg.Clout, type_ string) []plugin.Container {
	var containers []plugin.Container
	for _, vertex := range clout.Vertexes {
		containers = append(containers, GetContainers(vertex.Properties)...)
	}
	return containers
}
