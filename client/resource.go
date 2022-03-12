package client

import (
	contextpkg "context"
	"io"

	"github.com/tliron/khutulun/api"
)

type Resource struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Service   string `json:"service" yaml:"service"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
}

func (self *Client) ListResources(namespace string, serviceName string, type_ string) ([]Resource, error) {
	context, cancel := self.newContext()
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
			identifier, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return nil, err
				}
			}

			resources = append(resources, Resource{
				Namespace: identifier.Service.Namespace,
				Service:   identifier.Service.Name,
				Type:      identifier.Type,
				Name:      identifier.Name,
			})
		}

		return resources, nil
	} else {
		return nil, err
	}
}

func (self *Client) InteractRunnable(namespace string, serviceName string, resourceName string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	if client, err := self.client.InteractRunnable(contextpkg.Background()); err == nil {
		identifier := api.ResourceIdentifier{
			Service: &api.ServiceIdentifier{
				Namespace: namespace,
				Name:      serviceName,
			},
			Type: "runnable",
			Name: resourceName,
		}

		if err := client.Send(&api.Interaction{
			Resource: &identifier,
		}); err != nil {
			return err
		}

		go func() {
			var buffer []byte = make([]byte, 1)
			for {
				if _, err := stdin.Read(buffer); err == nil {
					if err := client.Send(&api.Interaction{
						Resource: &identifier,
						Stream:   "stdin",
						Bytes:    buffer,
					}); err != nil {
						log.Errorf("%s", err.Error())
						return
					}
				} else {
					if err != io.EOF {
						log.Errorf("%s", err.Error())
					}
					return
				}
			}
		}()

		for {
			interaction, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}

			switch interaction.Stream {
			case "stdout":
				if _, err := stdout.Write(interaction.Bytes); err != nil {
					return err
				}

			case "stderr":
				if _, err := stderr.Write(interaction.Bytes); err != nil {
					return err
				}
			}
		}

		return nil
	} else {
		return err
	}
}
