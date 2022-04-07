package client

import (
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

func (self *Client) Interact(identifier []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, pseudoTerminal bool, command ...string) error {
	if client, err := self.client.Interact(self.context); err == nil {
		if err := client.Send(&api.Interaction{
			Start: &api.Interaction_Start{
				Identifier:     identifier,
				Command:        command,
				PseudoTerminal: pseudoTerminal,
			},
		}); err != nil {
			return err
		}

		go func() {
			var buffer []byte = make([]byte, 1)
			for {
				if _, err := stdin.Read(buffer); err == nil {
					if err := client.Send(&api.Interaction{
						Stream: "stdin",
						Bytes:  buffer,
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
