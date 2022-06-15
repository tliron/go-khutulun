package sdk

import (
	"fmt"

	"github.com/tliron/kutil/ard"
	cloutpkg "github.com/tliron/puccini/clout"
	cloututil "github.com/tliron/puccini/clout/util"
)

//
// Connection
//

type Connection struct {
	Name string
	IP   string
	Port int64

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
	reflector := ard.NewReflector()
	reflector.IgnoreMissingStructFields = true
	reflector.NilMeansZero = true

	var relationship struct {
		Name       string `ard:"name"`
		Attributes struct {
			IP   string `ard:"ip"`
			Port int64  `ard:"port"`
		} `ard:"attributes"`
	}
	if err := reflector.ToComposite(edge.Properties, &relationship); err != nil {
		panic(err)
	}

	connection := Connection{
		Name:          fmt.Sprintf("%s:%d", relationship.Name, edgesOutIndex),
		IP:            relationship.Attributes.IP,
		Port:          relationship.Attributes.Port,
		vertexID:      vertex.ID,
		edgesOutIndex: edgesOutIndex,
	}

	if sources := GetVertexContainers(vertex); len(sources) > 0 {
		connection.Source = sources[0]
	}

	if edge.Target != nil {
		if targets := GetVertexContainers(edge.Target); len(targets) > 0 {
			connection.Target = targets[0]
		}
	}

	return connection
}

func GetVertexConnections(vertex *cloutpkg.Vertex) []Connection {
	var connections []Connection
	for index, edge := range cloututil.GetToscaRelationships(vertex, "cloud.puccini.khutulun::IPPort") {
		connections = append(connections, GetConnection(vertex, index, edge))
	}
	return connections
}

func GetCloutConnections(clout *cloutpkg.Clout) []Connection {
	var connections []Connection
	for _, vertex := range cloututil.GetToscaNodeTemplates(clout, "") {
		connections = append(connections, GetVertexConnections(vertex)...)
	}
	return connections
}
