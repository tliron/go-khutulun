package sdk

import (
	"fmt"

	"github.com/tliron/go-ard"
	cloutpkg "github.com/tliron/puccini/clout"
	cloututil "github.com/tliron/puccini/clout/util"
)

//
// Container
//

type OCIContainer struct {
	Capability

	Host            string
	Name            string
	Reference       string
	CreateArguments []string
	Ports           []OCIMappedPort
}

type OCIMappedPort struct {
	Address  string
	External int64
	Internal int64
	Protocol string
}

func GetOCIContainerMappedPorts(capability ard.Value) ([]OCIMappedPort, error) {
	var ports_ []struct {
		Properties struct {
			Internal int64  `ard:"internal"`
			Protocol string `ard:"protocol"`
		} `ard:"properties"`
		Attributes struct {
			Address  string `ard:"address"`
			External int64  `ard:"external"`
		} `ard:"attributes"`
	}
	if err := ardReflector.Pack(capability, &ports_); err != nil {
		return nil, err
	}

	ports := make([]OCIMappedPort, len(ports_))
	for index, port := range ports_ {
		ports[index] = OCIMappedPort{
			Address:  port.Attributes.Address,
			External: port.Attributes.External,
			Internal: port.Properties.Internal,
			Protocol: port.Properties.Protocol,
		}
	}

	return ports, nil
}

func GetOCIContainer(vertex *cloutpkg.Vertex, capability ard.Value, instanceName string, capabilityName string) (*OCIContainer, error) {
	var capability_ struct {
		Properties struct {
			Name            string            `ard:"name"`
			Image           OCIImageReference `ard:"image"`
			CreateArguments []string          `ard:"create-arguments"`
		} `ard:"properties"`
		Attributes struct {
			Host string `ard:"host"`
		} `ard:"attributes"`
	}
	if err := ardReflector.Pack(capability, &capability_); err != nil {
		return nil, err
	}

	if err := capability_.Properties.Image.Validate(); err != nil {
		return nil, err
	}

	if capability_.Properties.Name != "" {
		capabilityName = capability_.Properties.Name
	}

	return &OCIContainer{
		Host:            capability_.Attributes.Host,
		Name:            fmt.Sprintf("%s-%s", instanceName, capabilityName),
		Reference:       capability_.Properties.Image.String(),
		CreateArguments: capability_.Properties.CreateArguments,
		Capability: Capability{
			vertexID:       vertex.ID,
			capabilityName: capabilityName,
		},
	}, nil
}

func GetVertexOCIContainers(vertex *cloutpkg.Vertex) ([]*OCIContainer, error) {
	var instances []struct {
		Name string `ard:"name"`
	}
	if err := ardReflector.Pack(ard.With(vertex.Properties).Get("attributes", "instances").Value, &instances); err != nil {
		return nil, err
	}

	var containers []*OCIContainer

	for _, instance := range instances {
		instanceName := instance.Name

		var instanceContainers []*OCIContainer

		for capabilityName, capability := range cloututil.GetToscaCapabilities(vertex, "cloud.puccini.khutulun::OCIContainer") {
			if container, err := GetOCIContainer(vertex, capability, instanceName, capabilityName); err == nil {
				instanceContainers = append(instanceContainers, container)
			} else {
				return nil, err
			}
		}

		// All ports in node apply to all containers in node
		for _, capability := range cloututil.GetToscaCapabilities(vertex, "cloud.puccini.khutulun::MappedIPPort") {
			if ports, err := GetOCIContainerMappedPorts(capability); err == nil {
				for _, container := range instanceContainers {
					container.Ports = ports
				}
			} else {
				return nil, err
			}
		}

		containers = append(containers, instanceContainers...)
	}

	return containers, nil
}

func GetCloutOCIContainers(clout *cloutpkg.Clout) ([]*OCIContainer, error) {
	var containers []*OCIContainer
	for _, vertex := range cloututil.GetToscaNodeTemplates(clout, "cloud.puccini.khutulun::Instantiable") {
		if containers_, err := GetVertexOCIContainers(vertex); err == nil {
			containers = append(containers, containers_...)
		} else {
			return nil, err
		}
	}
	return containers, nil
}
