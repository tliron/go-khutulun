package sdk

import (
	"fmt"

	cloutpkg "github.com/tliron/puccini/clout"
	cloututil "github.com/tliron/puccini/clout/util"
)

//
// Connection
//

type Connection struct {
	Relationship

	Name   string
	IP     string
	Port   int64
	Source *Container
	Target *Container
}

func GetConnection(vertex *cloutpkg.Vertex, edgesOutIndex int, edge *cloutpkg.Edge) Connection {
	var relationship struct {
		Name       string `ard:"name"`
		Attributes struct {
			IP   string `ard:"ip"`
			Port int64  `ard:"port"`
		} `ard:"attributes"`
	}
	if err := ardReflector.ToComposite(edge.Properties, &relationship); err != nil {
		panic(err)
	}

	connection := Connection{
		Name: fmt.Sprintf("%s:%d", relationship.Name, edgesOutIndex),
		IP:   relationship.Attributes.IP,
		Port: relationship.Attributes.Port,
		Relationship: Relationship{
			vertexID:      vertex.ID,
			edgesOutIndex: edgesOutIndex,
		},
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
