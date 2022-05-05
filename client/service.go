package client

import (
	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/util"
)

func (self *Client) DeployService(serviceNamespace string, serviceName string, templateNamespace string, templateName string, inputs map[string]any) error {
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
	}

	context, cancel := self.newContextWithTimeout()
	defer cancel()

	_, err := self.client.DeployService(context, &deployService)
	return util.UnpackGRPCError(err)
}
