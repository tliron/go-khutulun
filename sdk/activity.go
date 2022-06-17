package sdk

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
)

//
// Capability
//

type Capability struct {
	vertexID       string
	capabilityName string
}

func (self *Capability) Find(clout *cloutpkg.Clout) (*cloutpkg.Vertex, ard.Value, error) {
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
