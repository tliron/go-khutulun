package client

import (
	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/util"
)

func (self *Client) DeployService(serviceNamespace string, serviceName string, templateNamespace string, templateName string, inputs map[string]any) error {
	args := api.DeployService{
		Template: &api.BundleIdentifier{
			Namespace: templateNamespace,
			Type:      &api.BundleType{Name: "template"},
			Name:      templateName,
		},
		Service: &api.ServiceIdentifier{
			Namespace: serviceNamespace,
			Name:      serviceName,
		},
	}

	context, cancel := self.newContextWithTimeout()
	defer cancel()

	_, err := self.client.DeployService(context, &args)
	return util.UnpackGrpcError(err)
}
