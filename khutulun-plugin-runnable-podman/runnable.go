package main

import (
	"fmt"
	"os"
	"os/exec"
	userpkg "os/user"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/tliron/kutil/ard"
)

const servicePrefix = "khutulun"

//
// Runnable
//

type Runnable struct{}

// systemctl --machine user@.host --user
// https://superuser.com/a/1461905

// plugin.Runnable interface
func (self *Runnable) Instantiate(config map[string]any) error {
	var config_ *Config
	var err error
	if config_, err = NewConfig(config); err != nil {
		return err
	}

	serviceName := fmt.Sprintf("%s-%s.service", servicePrefix, config_.name)

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

	args := []string{"create", "--name=" + config_.name, "--replace"} // --tty
	for _, port := range config_.ports {
		args = append(args, fmt.Sprintf("--publish=%d:%d/tcp", port.external, port.internal))
	}
	args = append(args, config_.source)
	args = append(args, config_.createArguments...)

	log.Infof("podman create %q", config_.name)
	command := exec.Command("podman", args...)
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "podman create")
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return errors.Wrap(err, "create systemd unit file")
	}
	defer file.Close()

	log.Infof("podman generate systemd %q", path)
	command = exec.Command("podman", "generate", "systemd", "--new", "--name", "--container-prefix="+servicePrefix, "--restart-policy=always", config_.name)
	command.Stdout = file
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "podman generate systemd")
	}

	command = exec.Command("systemctl", "--user", "daemon-reload")
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "systemctl daemon-reload")
	}

	log.Infof("systemctl enable %q", serviceName)
	command = exec.Command("systemctl", "--user", "enable", serviceName)
	if err := command.Run(); err != nil {
		return errors.Wrap(err, "systemctl enable")
	}

	log.Infof("systemctl restart %q", serviceName)
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

type Config struct {
	name            string
	source          string
	createArguments []string
	ports           []Port
}

type Port struct {
	external int64
	internal int64
}

func NewConfig(config map[string]any) (*Config, error) {
	var self Config

	config_ := ard.NewNode(config)
	var ok bool
	if self.name, ok = config_.Get("name").String(false); !ok {
		return nil, errors.New("\"name\" not provided")
	}
	if self.source, ok = config_.Get("source").String(false); !ok {
		return nil, errors.New("\"source\" not provided")
	}
	self.createArguments, _ = config_.Get("createArguments").StringList(false)
	if ports, ok := config_.Get("ports").List(false); ok {
		for _, port := range ports {
			var port_ Port
			port__ := ard.NewNode(port)
			if port_.external, ok = port__.Get("external").Integer(false); !ok {
				return nil, errors.New("\"port.external\" not provided")
			}
			if port_.internal, ok = port__.Get("internal").Integer(false); !ok {
				return nil, errors.New("\"port.internal\" not provided")
			}
		}
	}

	return &self, nil
}
