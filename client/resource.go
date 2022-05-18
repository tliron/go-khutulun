package client

import (
	"io"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/sdk"
)

type Resource struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Service   string `json:"service" yaml:"service"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
	Host      string `json:"host" yaml:"host"`
}

func (self *Client) ListResources(namespace string, serviceName string, type_ string) ([]Resource, error) {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	listResources := api.ListResources{
		Service: &api.ServiceIdentifier{
			Namespace: namespace,
			Name:      serviceName,
		},
		Type: type_,
	}

	if client, err := self.client.ListResources(context, &listResources); err == nil {
		var resources []Resource

		for {
			if identifier, err := client.Recv(); err == nil {
				resources = append(resources, Resource{
					Namespace: identifier.Service.Namespace,
					Service:   identifier.Service.Name,
					Type:      identifier.Type,
					Name:      identifier.Name,
					Host:      identifier.Host,
				})
			} else {
				if err == io.EOF {
					break
				} else {
					return nil, sdk.UnpackGRPCError(err)
				}
			}
		}

		return resources, nil
	} else {
		return nil, sdk.UnpackGRPCError(err)
	}
}
