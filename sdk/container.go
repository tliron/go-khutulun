package sdk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
	cloututil "github.com/tliron/puccini/clout/util"
)

//
// Container
//

type Container struct {
	Capability

	Host            string
	Name            string
	Reference       string
	CreateArguments []string
	Ports           []Port
}

type Port struct {
	External int64  `ard:"external"`
	Internal int64  `ard:"internal"`
	Protocol string `ard:"protocol"`
}

func GetContainerPorts(capability ard.Value) []Port {
	var ports []Port
	if ports_, ok := ard.NewNode(capability).Get("attributes", "ports").List(); ok {
		ard.NewReflector().ToComposite(ports_, &ports)
	}
	return ports
}

func GetContainers(vertex *cloutpkg.Vertex, capabilityName string, capability ard.Value) []*Container {
	var containers []*Container

	reflector := ard.NewReflector()
	reflector.IgnoreMissingStructFields = true
	reflector.NilMeansZero = true

	var instances []struct {
		Name string `ard:"name"`
	}
	if err := reflector.ToComposite(ard.NewNode(vertex.Properties).Get("attributes", "instances").Value, &instances); err != nil {
		panic(err)
	}

	var capability_ struct {
		Properties struct {
			Name            string                  `ard:"name"`
			Image           ContainerImageReference `ard:"image"`
			CreateArguments []string                `ard:"create-arguments"`
		} `ard:"properties"`
		Attributes struct {
			Host string `ard:"host"`
		} `ard:"attributes"`
	}
	if err := reflector.ToComposite(capability, &capability_); err != nil {
		panic(err)
	}

	for _, instance := range instances {
		container := Container{
			Host:            capability_.Attributes.Host,
			Name:            instance.Name,
			Reference:       capability_.Properties.Image.String(),
			CreateArguments: capability_.Properties.CreateArguments,
			Capability: Capability{
				vertexID:       vertex.ID,
				capabilityName: capabilityName,
			},
		}

		if capability_.Properties.Name != "" {
			container.Name = fmt.Sprintf("%s-%s", container.Name, capability_.Properties.Name)
		}

		containers = append(containers, &container)
	}

	return containers
}

func GetVertexContainers(vertex *cloutpkg.Vertex) []*Container {
	var containers []*Container
	if capabilities, ok := ard.NewNode(vertex.Properties).Get("capabilities").StringMap(); ok {
		for capabilityName, capability := range capabilities {
			if cloututil.IsToscaType(capability, "cloud.puccini.khutulun::Container") {
				containers = append(containers, GetContainers(vertex, capabilityName, capability)...)
			}
		}

		for _, capability := range capabilities {
			if cloututil.IsToscaType(capability, "cloud.puccini.khutulun::ContainerConnectable") {
				ports := GetContainerPorts(capability)
				for _, container := range containers {
					container.Ports = ports
				}
			}
		}
	}
	return containers
}

func GetCloutContainers(clout *cloutpkg.Clout) []*Container {
	var containers []*Container
	for _, vertex := range cloututil.GetToscaNodeTemplates(clout, "cloud.puccini.khutulun::Instantiated") {
		containers = append(containers, GetVertexContainers(vertex)...)
	}
	return containers
}

//
// ContainerImageReference
//

type ContainerImageReference struct {
	Reference string `ard:"reference"`

	Host            string `ard:"host"`
	Port            int    `ard:"port"`
	Repository      string `ard:"repository"`
	Image           string `ard:"image"`
	Tag             string `ard:"tag"`
	DigestAlgorithm string `ard:"digestAlgorithm"`
	DigestHex       string `ard:"digestHex"`
}

// [host[:port]/][repository/]image[:tag][@digest-algorithm:digest-hex]
// fmt.Stringer interface
func (self ContainerImageReference) String() string {
	if self.Reference != "" {
		return self.Reference
	}

	var s strings.Builder
	if self.Host != "" {
		s.WriteString(self.Host)
		if self.Port != 0 {
			s.WriteRune(':')
			s.WriteString(strconv.Itoa(self.Port))
		}
		s.WriteRune('/')
	}
	if self.Repository != "" {
		s.WriteString(self.Repository)
		s.WriteRune('/')
	}
	s.WriteString(self.Image)
	if self.Tag != "" {
		s.WriteRune(':')
		s.WriteString(self.Tag)
	}
	if self.DigestAlgorithm != "" {
		s.WriteRune('@')
		s.WriteString(self.DigestAlgorithm)
		if self.DigestHex != "" {
			s.WriteRune(':')
			s.WriteString(self.DigestHex)
		}
	}
	return s.String()
}
