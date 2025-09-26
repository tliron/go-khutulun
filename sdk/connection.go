package sdk

import (
	"fmt"

	cloutpkg "github.com/tliron/go-puccini/clout"
	cloututil "github.com/tliron/go-puccini/clout/util"
)

//
// Connection
//

type Connection struct {
	Relationship

	Name   string
	IP     string
	Port   int64
	Source *OCIContainer
	Target *OCIContainer
}

func GetConnection(vertex *cloutpkg.Vertex, edgesOutIndex int, edge *cloutpkg.Edge) (*Connection, error) {
	var relationship struct {
		Name       string `ard:"name"`
		Attributes struct {
			IP   string `ard:"ip"`
			Port int64  `ard:"port"`
		} `ard:"attributes"`
	}
	if err := ardReflector.Pack(edge.Properties, &relationship); err != nil {
		return nil, err
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

	if sources, err := GetVertexOCIContainers(vertex); err == nil {
		if len(sources) > 0 {
			connection.Source = sources[0]
		} else {
			return nil, err
		}
	}

	if edge.Target != nil {
		if targets, err := GetVertexOCIContainers(edge.Target); err == nil {
			if len(targets) > 0 {
				connection.Target = targets[0]
			}
		} else {
			return nil, err
		}
	}

	return &connection, nil
}

func GetVertexConnections(vertex *cloutpkg.Vertex) ([]*Connection, error) {
	var connections []*Connection
	for index, edge := range cloututil.GetToscaRelationships(vertex, "cloud.puccini.khutulun::IPPort") {
		if connection, err := GetConnection(vertex, index, edge); err == nil {
			connections = append(connections, connection)
		} else {
			return nil, err
		}
	}
	return connections, nil
}

func GetCloutConnections(clout *cloutpkg.Clout) ([]*Connection, error) {
	var connections []*Connection
	for _, vertex := range cloututil.GetToscaNodeTemplates(clout, "") {
		if connections_, err := GetVertexConnections(vertex); err == nil {
			connections = append(connections, connections_...)
		} else {
			return nil, err
		}
	}
	return connections, nil
}
