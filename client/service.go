package client

import (
	"github.com/tliron/go-khutulun/api"
	"github.com/tliron/go-khutulun/sdk"
)

func (self *Client) DeployService(serviceNamespace string, serviceName string, templateNamespace string, templateName string, inputs map[string]any, async bool) error {
	deployService := api.DeployService{
		Template: &api.PackageIdentifier{
			Namespace: templateNamespace,
			Type:      &api.PackageType{Name: "template"},
			Name:      templateName,
		},
		Service: &api.ServiceIdentifier{
			Namespace: serviceNamespace,
			Name:      serviceName,
		},
		Async: async,
	}

	context, cancel := self.newContextWithTimeout()
	defer cancel()

	_, err := self.client.DeployService(context, &deployService)
	return sdk.UnpackGRPCError(err)
}
