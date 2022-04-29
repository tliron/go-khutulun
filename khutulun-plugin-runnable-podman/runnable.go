package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/khutulun/util"
	"github.com/tliron/kutil/protobuf"
	utilpkg "github.com/tliron/kutil/util"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

const servicePrefix = "khutulun"

//
// Runnable
//

type Runnable struct{}

// systemctl --machine user@.host --user
// https://superuser.com/a/1461905

// plugin.Runnable interface
func (self *Runnable) Instantiate(config any) error {
	var container plugin.Container
	if err := protobuf.UnpackStringMap(config, &container); err != nil {
		return err
	}

	serviceName := fmt.Sprintf("%s-%s.service", servicePrefix, container.Name)

	user_, err := user.Current()
	if err != nil {
		return fmt.Errorf("current user: %w", err)
	}

	path := filepath.Join(user_.HomeDir, ".config", "systemd", "user", serviceName)
	if exists, err := utilpkg.FileExists(path); err == nil {
		if exists {
			log.Infof("systemd unit already exists: %q", path)
			return nil
		} else {
			log.Infof("systemd unit: %q", path)
		}
	} else {
		return fmt.Errorf("file: %w", err)
	}

	args := []string{"create", "--name=" + container.Name, "--replace"} // --tty?
	args = append(args, container.CreateArguments...)
	for _, port := range container.Ports {
		args = append(args, fmt.Sprintf("--publish=%d:%d/%s", port.External, port.Internal, port.Protocol))
	}
	args = append(args, container.Reference)

	log.Infof("podman %s", strings.Join(args, " "))
	command := exec.Command("/usr/bin/podman", args...)
	if err := command.Run(); err != nil {
		return fmt.Errorf("podman create: %w", err)
	}

	log.Infof("mkdir --parents %q", path)
	command = exec.Command("/usr/bin/mkdir", "--parents", filepath.Dir(path))
	if err := command.Run(); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("create systemd unit file: %w", err)
	}
	defer file.Close()

	args = []string{"generate", "systemd", "--new", "--name", "--container-prefix=" + servicePrefix, "--restart-policy=always", container.Name}
	log.Infof("podman %s", strings.Join(args, " "))
	command = exec.Command("/usr/bin/podman", args...)
	command.Stdout = file
	if err := command.Run(); err != nil {
		return fmt.Errorf("podman generate systemd: %w", err)
	}

	command = exec.Command("/usr/bin/systemctl", "--user", "daemon-reload")
	if err := command.Run(); err != nil {
		return fmt.Errorf("systemctl daemon-reload: %w", err)
	}

	log.Infof("systemctl enable %s", serviceName)
	command = exec.Command("/usr/bin/systemctl", "--user", "enable", serviceName)
	if err := command.Run(); err != nil {
		return fmt.Errorf("systemctl enable: %w", err)
	}

	log.Infof("systemctl restart %s", serviceName)
	command = exec.Command("/usr/bin/systemctl", "--user", "--no-block", "restart", serviceName)
	if err := command.Run(); err != nil {
		return fmt.Errorf("systemctl start: %w", err)
	}

	log.Info("loginctl enable-linger")
	command = exec.Command("/usr/bin/loginctl", "enable-linger")
	if err := command.Run(); err != nil {
		return fmt.Errorf("loginctl enable-linger: %w", err)
	}

	return nil
}

// plugin.Runnable interface
func (self *Runnable) Interact(server util.Interactor, start *api.Interaction_Start) error {
	if len(start.Identifier) != 4 {
		return statuspkg.Errorf(codes.InvalidArgument, "malformed identifier for runnable: %s", start.Identifier)
	}

	//namespace := interaction.Start.Identifier[1]
	//serviceName := interaction.Start.Identifier[2]
	resourceName := start.Identifier[3]

	command := util.NewCommand(start, log)
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

	return util.StartCommand(command, server, log)
}
