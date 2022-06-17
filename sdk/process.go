package sdk

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
	cloututil "github.com/tliron/puccini/clout/util"
)

//
// Process
//

type Process struct {
	Capability

	Host      string
	Name      string
	Command   string
	Arguments []string
}

func GetProcesses(vertex *cloutpkg.Vertex, capabilityName string, capability ard.Value) []*Process {
	var processes []*Process

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
			Name    string `ard:"name"`
			Command struct {
				Name      string   `ard:"name"`
				Arguments []string `ard:"arguments"`
			} `ard:"command"`
		} `ard:"properties"`
		Attributes struct {
			Host string `ard:"host"`
		} `ard:"attributes"`
	}
	if err := reflector.ToComposite(capability, &capability_); err != nil {
		panic(err)
	}

	for _, instance := range instances {
		process := Process{
			Host:      capability_.Attributes.Host,
			Name:      instance.Name,
			Command:   capability_.Properties.Command.Name,
			Arguments: capability_.Properties.Command.Arguments,
			Capability: Capability{
				vertexID:       vertex.ID,
				capabilityName: capabilityName,
			},
		}

		if capability_.Properties.Name != "" {
			process.Name = fmt.Sprintf("%s-%s", process.Name, capability_.Properties.Name)
		}

		processes = append(processes, &process)
	}

	return processes
}

func GetVertexProcesses(vertex *cloutpkg.Vertex) []*Process {
	var processes []*Process
	if capabilities, ok := ard.NewNode(vertex.Properties).Get("capabilities").StringMap(); ok {
		for capabilityName, capability := range capabilities {
			if cloututil.IsToscaType(capability, "cloud.puccini.khutulun::Process") {
				processes = append(processes, GetProcesses(vertex, capabilityName, capability)...)
			}
		}
	}
	return processes
}

func GetCloutProcesses(clout *cloutpkg.Clout) []*Process {
	var processes []*Process
	for _, vertex := range cloututil.GetToscaNodeTemplates(clout, "cloud.puccini.khutulun::Instantiated") {
		processes = append(processes, GetVertexProcesses(vertex)...)
	}
	return processes
}
