package agent

import (
	"fmt"

	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
)

//
// Connection
//

type Connection struct {
	plugin.Connection

	Source *Container
	Target *Container

	vertexID      string
	edgesOutIndex int
}

func (self *Connection) Find(clout *cloutpkg.Clout) (*cloutpkg.Edge, error) {
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

func GetConnection(vertex *cloutpkg.Vertex, edgesOutIndex int, edge *cloutpkg.Edge) Connection {
	self := Connection{
		vertexID:      vertex.ID,
		edgesOutIndex: edgesOutIndex,
	}

	name, _ := ard.NewNode(edge.Properties).Get("name").String()
	self.Name = fmt.Sprintf("%s:%d", name, edgesOutIndex)
	relationshipAttributes, _ := ard.NewNode(edge.Properties).Get("attributes").StringMap()
	self.IP, _ = ard.NewNode(relationshipAttributes).Get("ip").String()
	port, _ := ard.NewNode(relationshipAttributes).Get("port").Integer()
	self.Port = int(port)

	if sources := GetVertexContainers(vertex); len(sources) > 0 {
		self.Source = sources[0]
	}

	if edge.Target != nil {
		if targets := GetVertexContainers(edge.Target); len(targets) > 0 {
			self.Target = targets[0]
		}
	}

	return self
}

func GetVertexConnections(vertex *cloutpkg.Vertex) []Connection {
	var connections []Connection
	for index, edge := range vertex.EdgesOut {
		if types, ok := ard.NewNode(edge.Properties).Get("types").StringMap(); ok {
			if _, ok := types["cloud.puccini.khutulun::IPPort"]; ok {
				connections = append(connections, GetConnection(vertex, index, edge))
			}
		}
	}
	return connections
}

func GetCloutConnections(clout *cloutpkg.Clout) []Connection {
	var connections []Connection
	for _, vertex := range clout.Vertexes {
		connections = append(connections, GetVertexConnections(vertex)...)
	}
	return connections
}
