package main

import (
	"fmt"
	"os/exec"

	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/util"
	cloutpkg "github.com/tliron/puccini/clout"
)

// https://docs.podman.io/en/latest/markdown/podman-create.1.html
// https://docs.podman.io/en/latest/markdown/podman-generate-systemd.1.html

// systemctl --machine user@.host --user
// https://superuser.com/a/1461905

func (self *Delegate) Reconcile(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	containers, err := sdk.GetCloutOCIContainers(coercedClout)
	if err != nil {
		return nil, nil, err
	}
	if len(containers) == 0 {
		return nil, nil, nil
	}

	for _, container := range containers {
		if container.Host == self.host {
			//format.WriteGo(container, logging.GetWriter(), " ")
			if err := self.CreateContainerUserService(container); err != nil {
				log.Errorf("instantiate: %s", err.Error())
			}
		}
	}

	return nil, nil, nil
}

func (self *Delegate) CreateContainerUserService(container *sdk.OCIContainer) error {
	serviceName := fmt.Sprintf("%s-%s.service", servicePrefix, container.Name)

	file, err := sdk.CreateUserSystemdFile(serviceName, log)
	if err != nil {
		return err
	}

	args := []string{"create", "--name=" + container.Name, "--replace"} // --tty?
	args = append(args, container.CreateArguments...)
	for _, port := range container.Ports {
		if (port.Address != "") && (port.External != 0) && (port.Internal != 0) {
			var protocol string
			switch port.Protocol {
			case "UDP":
				protocol = "udp"
			case "SCTP":
				protocol = "sctp"
			default:
				protocol = "tcp"
			}
			args = append(args, fmt.Sprintf("--publish=%s:%d:%d/%s", port.Address, port.External, port.Internal, protocol))
		}
	}
	args = append(args, container.Reference)

	log.Infof("podman %s", util.JoinQuote(args, " "))
	command := exec.Command("/usr/bin/podman", args...)
	if err := command.Run(); err != nil {
		file.Close()
		return fmt.Errorf("podman create: %w", err)
	}

	args = []string{"generate", "systemd", "--new", "--name", "--container-prefix=" + servicePrefix, "--restart-policy=always", container.Name}
	log.Infof("podman %s", util.JoinQuote(args, " "))
	command = exec.Command("/usr/bin/podman", args...)
	command.Stdout = file
	if err := command.Run(); err != nil {
		file.Close()
		return fmt.Errorf("podman generate systemd: %w", err)
	}

	file.Close()
	return sdk.EnableUserSystemd(serviceName, log)
}
