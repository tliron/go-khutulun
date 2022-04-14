package client

import (
	"io"

	"github.com/tliron/khutulun/api"
)

const BUFFER_SIZE = 65536

type Artifact struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
}

func (self *Client) ListArtifacts(namespace string, type_ string) ([]Artifact, error) {
	context, cancel := self.newContext()
	defer cancel()

	listArtifacts := api.ListArtifacts{
		Namespace: namespace,
		Type:      &api.ArtifactType{Name: type_},
	}

	if client, err := self.client.ListArtifacts(context, &listArtifacts); err == nil {
		var artifacts []Artifact

		for {
			identifier, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return nil, err
				}
			}

			artifacts = append(artifacts, Artifact{
				Namespace: identifier.Namespace,
				Type:      identifier.Type.Name,
				Name:      identifier.Name,
			})
		}

		return artifacts, nil
	} else {
		return nil, err
	}
}

func (self *Client) GetArtifact(namespace string, type_ string, name string, writer io.Writer) error {
	context, cancel := self.newContext()
	defer cancel()

	identifier := api.ArtifactIdentifier{
		Namespace: namespace,
		Type:      &api.ArtifactType{Name: type_},
		Name:      name,
	}

	if client, err := self.client.GetArtifact(context, &identifier); err == nil {
		for {
			content, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}

			switch content_ := content.Content.(type) {
			case *api.ArtifactContent_Bytes:
				if _, err := writer.Write(content_.Bytes); err != nil {
					return err
				}
			}
		}

		return nil
	} else {
		return err
	}
}

func (self *Client) SetArtifact(namespace string, type_ string, name string, reader io.Reader) error {
	context, cancel := self.newContext()
	defer cancel()

	if client, err := self.client.SetArtifact(context); err == nil {
		identifier := api.ArtifactIdentifier{
			Namespace: namespace,
			Type:      &api.ArtifactType{Name: type_},
			Name:      name,
		}

		content := api.ArtifactContent{Content: &api.ArtifactContent_Identifier{Identifier: &identifier}}
		if err := client.Send(&content); err != nil {
			return err
		}

		buffer := make([]byte, BUFFER_SIZE)
		for {
			if count, err := reader.Read(buffer); err == nil {
				content := api.ArtifactContent{Content: &api.ArtifactContent_Bytes{Bytes: buffer[:count]}}
				if err := client.Send(&content); err != nil {
					return err
				}
			} else {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}
		}

		if _, err := client.CloseAndRecv(); err == io.EOF {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *Client) RemoveArtifact(namespace string, type_ string, name string) error {
	context, cancel := self.newContext()
	defer cancel()

	identifier := api.ArtifactIdentifier{
		Namespace: namespace,
		Type:      &api.ArtifactType{Name: type_},
		Name:      name,
	}

	_, err := self.client.RemoveArtifact(context, &identifier)
	return err
}
