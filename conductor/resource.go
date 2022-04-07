package conductor

import (
	"errors"
	"fmt"
	"os"

	"github.com/tliron/kutil/exec"
)

type ResourceIdentifier struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Service   string `json:"service" yaml:"service"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
}

func (self *Conductor) ListResources(namespace string, serviceName string, type_ string) ([]ResourceIdentifier, error) {
	var resources []ResourceIdentifier

	var artifacts []ArtifactIdentifier
	if serviceName == "" {
		var err error
		if artifacts, err = self.ListArtifacts(namespace, "clout"); err != nil {
			return nil, err
		}
	} else {
		artifacts = []ArtifactIdentifier{
			{
				Namespace: namespace,
				Type:      "clout",
				Name:      serviceName,
			},
		}
	}

	for _, artifact := range artifacts {
		if clout, err := self.GetClout(artifact.Namespace, artifact.Name, true); err == nil {
			for _, resource := range self.getResources(clout, type_) {
				resources = append(resources, ResourceIdentifier{
					Namespace: artifact.Namespace,
					Service:   artifact.Name,
					Type:      type_,
					Name:      resource.Name,
				})
			}
		} else {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}
	}

	return resources, nil
}

// Caller has to close stdin, otherwise there will be a goroutine leak!
func (self *Conductor) Interact(identifier []string, pseudoTerminal bool, done chan error, command ...string) (chan struct{}, chan []byte, chan []byte, chan []byte, error) {
	if len(identifier) == 0 {
		return nil, nil, nil, nil, errors.New("no identifier")
	}
	type_ := identifier[0]

	if len(command) == 0 {
		command = []string{"/bin/bash"}
		if pseudoTerminal {
			// We need to force interactive mode for bash
			command = append(command, "-i")
		}
	}

	var name string
	var args []string

	switch type_ {
	case "host":
		name = command[0]
		if len(command) > 1 {
			args = command[1:]
		}

	case "runnable":
		if len(identifier) != 4 {
			return nil, nil, nil, nil, fmt.Errorf("malformed identifier for runnable: %s", identifier)
		}

		//namespace := identifier[1]
		//serviceName := identifier[2]
		resourceName := identifier[3]

		name = "podman"
		args = []string{"exec"}
		if pseudoTerminal {
			args = append(args, "--interactive", "--tty")
			pseudoTerminal = false // no need to nest a pseudo-terminal
		}
		args = append(args, resourceName)
		args = append(args, command...)

	default:
		return nil, nil, nil, nil, fmt.Errorf("unsupported identifier type: %s", type_)
	}

	return exec.ExecInteractive(pseudoTerminal, done, name, args...)
}
