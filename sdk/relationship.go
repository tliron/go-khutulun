package sdk

import (
	"fmt"

	cloutpkg "github.com/tliron/go-puccini/clout"
)

//
// Relationship
//

type Relationship struct {
	vertexID      string
	edgesOutIndex int
}

func (self *Relationship) Find(clout *cloutpkg.Clout) (*cloutpkg.Edge, error) {
	if vertex, ok := clout.Vertexes[self.vertexID]; ok {
		if self.edgesOutIndex < len(vertex.EdgesOut) {
			return vertex.EdgesOut[self.edgesOutIndex], nil
		} else {
			return nil, fmt.Errorf("vertex has too few edges: %s", self.vertexID)
		}
	} else {
		return nil, fmt.Errorf("vertex not found: %s", self.vertexID)
	}
}
