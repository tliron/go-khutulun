package main

import (
	"fmt"
	"os"
	"os/exec"
	userpkg "os/user"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/kutil/protobuf"
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

	user, err := userpkg.Current()
	if err != nil {
		return errors.Wrap(err, "current user")
	}

	path := filepath.Join(user.HomeDir, ".config", "systemd", "user", serviceName)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Infof("systemd unit: %q", path)
	} else {
		log.Infof("systemd unit already exists: %q", path)
		return nil
	}

	args := []string{"create", "--name=" + container.Name, "--replace"} // --tty?
	args = append(args, container.CreateArguments...)
	for _, port := range container.Ports {
		args = append(args, fmt.Sprintf("--publish=%d:%d/tcp", port.External, port.Internal))
	}
	args = append(args, container.Reference)

	log.Infof("podman %s", strings.Join(args, " "))
	command := exec.Command("podman", args...)
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "podman create")
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return errors.Wrap(err, "create systemd unit file")
	}
	defer file.Close()

	args = []string{"generate", "systemd", "--new", "--name", "--container-prefix=" + servicePrefix, "--restart-policy=always", container.Name}
	log.Infof("podman %s", strings.Join(args, " "))
	command = exec.Command("podman", args...)
	command.Stdout = file
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "podman generate systemd")
	}

	command = exec.Command("systemctl", "--user", "daemon-reload")
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "systemctl daemon-reload")
	}

	log.Infof("systemctl enable %s", serviceName)
	command = exec.Command("systemctl", "--user", "enable", serviceName)
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "systemctl enable")
	}

	log.Infof("systemctl restart %s", serviceName)
	command = exec.Command("systemctl", "--user", "--no-block", "restart", serviceName)
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "systemctl start")
	}

	log.Info("loginctl enable-linger")
	command = exec.Command("loginctl", "enable-linger")
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "loginctl enable-linger")
	}

	return nil
}
