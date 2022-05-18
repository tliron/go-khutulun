package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/util"
	cloutpkg "github.com/tliron/puccini/clout"
)

// systemctl --machine user@.host --user
// https://superuser.com/a/1461905

func (self *Delegate) Reconcile(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, error) {
	containers := sdk.GetCloutContainers(clout)
	if len(containers) == 0 {
		return nil, nil
	}

	for _, container := range containers {
		if container.Host == self.host {
			if err := self.CreateContainer(container); err != nil {
				log.Errorf("instantiate: %s", err.Error())
			}
		}
	}

	return nil, nil
}

func (self *Delegate) CreateContainer(container *sdk.Container) error {
	serviceName := fmt.Sprintf("%s-%s.service", servicePrefix, container.Name)

	user_, err := user.Current()
	if err != nil {
		return fmt.Errorf("current user: %w", err)
	}

	path := filepath.Join(user_.HomeDir, ".config", "systemd", "user", serviceName)
	if exists, err := util.DoesFileExist(path); err == nil {
		if exists {
			log.Infof("systemd unit already exists: %q", path)
			//return nil
		} else {
			log.Infof("systemd unit: %q", path)
		}
	} else {
		return fmt.Errorf("file: %w", err)
	}

	args := []string{"create", "--name=" + container.Name, "--replace"} // --tty?
	args = append(args, container.CreateArguments...)
	for _, port := range container.Ports {
		protocol := "tcp"
		switch port.Protocol {
		case "UDP":
			protocol = "udp"
		case "SCTP":
			protocol = "sctp"
		}
		args = append(args, fmt.Sprintf("--publish=%d:%d/%s", port.External, port.Internal, protocol))
	}
	args = append(args, container.Reference)

	log.Infof("podman %s", util.Joinq(args, " "))
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
	log.Infof("podman %s", util.Joinq(args, " "))
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

//
// Reconcile
//

type Reconcile map[string]*ServiceIdentifiers

func NewReconcile() Reconcile {
	return make(map[string]*ServiceIdentifiers)
}

func (self Reconcile) Add(host string, identifier *ServiceIdentifier) bool {
	var identifiers *ServiceIdentifiers
	var ok bool
	if identifiers, ok = self[host]; !ok {
		identifiers = NewServiceIdentifiers()
		self[host] = identifiers
	}
	return identifiers.Add(identifier)
}

func (self Reconcile) Merge(reconcile Reconcile) bool {
	var added bool
	for host, identifiers := range reconcile {
		for _, identifier := range identifiers.List {
			if self.Add(host, identifier) {
				added = true
			}
		}
	}
	return added
}
