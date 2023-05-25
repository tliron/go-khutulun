package agent

import (
	contextpkg "context"
	"os"
	"sort"
	"strings"

	"github.com/tliron/commonlog"
	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

type ResourceIdentifier struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Service   string `json:"service" yaml:"service"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
	Host      string `json:"host" yaml:"host"`
}

type ResourceIdentifiers []ResourceIdentifier

// sort.Interface interface
func (self ResourceIdentifiers) Len() int {
	return len(self)
}

// sort.Interface interface
func (self ResourceIdentifiers) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

// sort.Interface interface
func (self ResourceIdentifiers) Less(i, j int) bool {
	if c := strings.Compare(self[i].Namespace, self[j].Namespace); c == 0 {
		if c := strings.Compare(self[i].Service, self[j].Service); c == 0 {
			if c := strings.Compare(self[i].Type, self[j].Type); c == 0 {
				return strings.Compare(self[i].Name, self[j].Name) == -1
			} else {
				return c == 1
			}
		} else {
			return c == 1
		}
	} else {
		return c == -1
	}
}

func (self *Agent) ListResources(context contextpkg.Context, namespace string, serviceName string, type_ string) (ResourceIdentifiers, error) {
	var resources ResourceIdentifiers

	var packages []sdk.PackageIdentifier
	if serviceName == "" {
		var err error
		if packages, err = self.state.ListPackages(namespace, "service"); err != nil {
			return nil, err
		}
	} else {
		packages = []sdk.PackageIdentifier{
			{
				Namespace: namespace,
				Type:      "service",
				Name:      serviceName,
			},
		}
	}

	for _, package_ := range packages {
		if lock, clout, err := self.state.OpenServiceClout(context, package_.Namespace, package_.Name, self.urlContext); err == nil {
			commonlog.CallAndLogError(lock.Unlock, "unlock", log)
			if clout, err = self.CoerceClout(clout, false); err == nil {
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

	sort.Sort(resources)

	return resources, nil
}

func (self *Agent) getResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout, type_ string) ([]delegatepkg.Resource, error) {
	delegates := self.NewDelegates()
	defer delegates.Release()
	delegates.Fill(namespace, coercedClout)

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
