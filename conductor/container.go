package conductor

import (
	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/kutil/ard"
)

func GetContainer(vertexProperties any, capabilityProperties any) plugin.Container {
	var self plugin.Container
	var ok bool
	if self.Name, ok = ard.NewNode(capabilityProperties).Get("name").String(); !ok {
		self.Name, _ = ard.NewNode(vertexProperties).Get("name").String()
	}
	self.Reference, _ = ard.NewNode(capabilityProperties).Get("image").Get("reference").String()
	self.CreateArguments, _ = ard.NewNode(capabilityProperties).Get("create-arguments").StringList()
	if ports, ok := ard.NewNode(capabilityProperties).Get("ports").List(); ok {
		for _, port := range ports {
			external, _ := ard.NewNode(port).Get("external").Integer()
			internal, _ := ard.NewNode(port).Get("internal").Integer()
			protocol, _ := ard.NewNode(port).Get("protocol").String()
			self.Ports = append(self.Ports, plugin.Port{
				External: external,
				Internal: internal,
				Protocol: protocol,
			})
		}
	}
	return self
}

func GetContainers(vertexProperties any) []plugin.Container {
	var containers []plugin.Container
	if capabilities, ok := ard.NewNode(vertexProperties).Get("capabilities").StringMap(); ok {
		for _, capability := range capabilities {
			if types, ok := ard.NewNode(capability).Get("types").StringMap(); ok {
				if _, ok := types["cloud.puccini.khutulun::Container"]; ok {
					capabilityProperties, _ := ard.NewNode(capability).Get("properties").StringMap()
					containers = append(containers, GetContainer(vertexProperties, capabilityProperties))
				}
			}
		}
	}
	return containers
}
