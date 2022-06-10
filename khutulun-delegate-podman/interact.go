package main

import (
	"fmt"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/sdk"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

// delegate.Delegate interface
func (self *Delegate) Interact(server sdk.GRPCInteractor, start *api.Interaction_Start) error {
	if len(start.Identifier) != 4 {
		return statuspkg.Errorf(codes.InvalidArgument, "malformed identifier for activity: %s", start.Identifier)
	}

	//namespace := interaction.Start.Identifier[1]
	//serviceName := interaction.Start.Identifier[2]
	resourceName := start.Identifier[3]

	command := sdk.NewCommand(start, log)
	args := append([]string{command.Name}, command.Args...)
	command.Name = "/usr/bin/podman"
	command.Args = []string{"exec"}

	if command.PseudoTerminal != nil {
		command.Args = append(command.Args, "--interactive", "--tty")
	}

	if command.Environment != nil {
		for k, v := range command.Environment {
			command.Args = append(command.Args, fmt.Sprintf("--env=%s=%s", k, v))
			delete(command.Environment, k)
		}
	}

	// Needed for podman to access "nsenter"
	command.AddPath("PATH", "/usr/bin")

	command.Args = append(command.Args, resourceName)
	command.Args = append(command.Args, args...)

	return sdk.StartCommand(command, server, log)
}
