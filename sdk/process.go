package sdk

import (
	"fmt"

	"github.com/tliron/go-ard"
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

func GetProcesses(vertex *cloutpkg.Vertex, capabilityName string, capability ard.Value) ([]*Process, error) {
	var processes []*Process

	var instances []struct {
		Name string `ard:"name"`
	}
	if err := ardReflector.Pack(ard.NewNode(vertex.Properties).Get("attributes", "instances").Value, &instances); err != nil {
		return nil, err
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
	if err := ardReflector.Pack(capability, &capability_); err != nil {
		return nil, err
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

	return processes, nil
}

func GetVertexProcesses(vertex *cloutpkg.Vertex) ([]*Process, error) {
	var processes []*Process
	for capabilityName, capability := range cloututil.GetToscaCapabilities(vertex, "cloud.puccini.khutulun::Process") {
		if processes_, err := GetProcesses(vertex, capabilityName, capability); err == nil {
			processes = append(processes, processes_...)
		} else {
			return nil, err
		}
	}
	return processes, nil
}

func GetCloutProcesses(clout *cloutpkg.Clout) ([]*Process, error) {
	var processes []*Process
	for _, vertex := range cloututil.GetToscaNodeTemplates(clout, "cloud.puccini.khutulun::Instantiated") {
		if processes_, err := GetVertexProcesses(vertex); err == nil {
			processes = append(processes, processes_...)
		} else {
			return nil, err
		}
	}
	return processes, nil
}
