package host

import (
	"fmt"

	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
)

//
// Container
//

type Container struct {
	plugin.Container

	vertexID       string
	capabilityName string
}

func (self *Container) Find(clout *cloutpkg.Clout) (*cloutpkg.Vertex, any, error) {
	if vertex, ok := clout.Vertexes[self.vertexID]; ok {
		if capabilities, ok := ard.NewNode(vertex.Properties).Get("capabilities").StringMap(); ok {
			if capability, ok := capabilities[self.capabilityName]; ok {
				return vertex, capability, nil
			} else {
				return nil, nil, fmt.Errorf("vertex %s has no capability: %s", self.vertexID, self.capabilityName)
			}
		} else {
			return nil, nil, fmt.Errorf("vertex has no capabilities: %s", self.vertexID)
		}
	} else {
		return nil, nil, fmt.Errorf("vertex not found: %s", self.vertexID)
	}
}

func GetContainerPorts(capability any) []plugin.Port {
	var ports []plugin.Port
	capabilityProperties, _ := ard.NewNode(capability).Get("properties").StringMap()
	if ports_, ok := ard.NewNode(capabilityProperties).Get("ports").List(); ok {
		for _, port := range ports_ {
			external, _ := ard.NewNode(port).Get("external").Integer()
			internal, _ := ard.NewNode(port).Get("internal").Integer()
			protocol, _ := ard.NewNode(port).Get("protocol").String()
			ports = append(ports, plugin.Port{
				External: external,
				Internal: internal,
				Protocol: protocol,
			})
		}
	}
	return ports
}

func GetContainer(vertex *cloutpkg.Vertex, capabilityName string, capability any) *Container {
	self := Container{
		vertexID:       vertex.ID,
		capabilityName: capabilityName,
	}

	capabilityProperties, _ := ard.NewNode(capability).Get("properties").StringMap()
	capabilityAttributes, _ := ard.NewNode(capability).Get("attributes").StringMap()

	self.Host, _ = ard.NewNode(capabilityAttributes).Get("host").String()
	var ok bool
	if self.Name, ok = ard.NewNode(capabilityProperties).Get("name").String(); !ok {
		self.Name, _ = ard.NewNode(vertex.Properties).Get("name").String()
	}
	self.Reference, _ = ard.NewNode(capabilityProperties).Get("image").Get("reference").String()
	self.CreateArguments, _ = ard.NewNode(capabilityProperties).Get("create-arguments").StringList()

	return &self
}

func GetVertexContainers(vertex *cloutpkg.Vertex) []*Container {
	var containers []*Container
	if capabilities, ok := ard.NewNode(vertex.Properties).Get("capabilities").StringMap(); ok {
		for capabilityName, capability := range capabilities {
			if types, ok := ard.NewNode(capability).Get("types").StringMap(); ok {
				if _, ok := types["cloud.puccini.khutulun::Container"]; ok {
					containers = append(containers, GetContainer(vertex, capabilityName, capability))
				}
			}
		}

		for _, capability := range capabilities {
			if types, ok := ard.NewNode(capability).Get("types").StringMap(); ok {
				if _, ok := types["cloud.puccini.khutulun::ContainerConnectable"]; ok {
					ports := GetContainerPorts(capability)
					for _, container := range containers {
						container.Ports = ports
					}
				}
			}
		}
	}
	return containers
}

func GetCloutContainers(clout *cloutpkg.Clout) []*Container {
	var containers []*Container
	for _, vertex := range clout.Vertexes {
		containers = append(containers, GetVertexContainers(vertex)...)
	}
	return containers
}
