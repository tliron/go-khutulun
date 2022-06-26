package sdk

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
)

const servicePrefix = "khutulun"

func CreateUserSystemdFile(name string, log logging.Logger) (*os.File, error) {
	user_, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("current user: %w", err)
	}

	path := filepath.Join(user_.HomeDir, ".config", "systemd", "user", name)
	if exists, err := util.DoesFileExist(path); err == nil {
		if exists {
			log.Infof("systemd unit already exists: %q", path)
			//return nil
		} else {
			log.Infof("systemd unit: %q", path)
		}
	} else {
		return nil, fmt.Errorf("file: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("create systemd unit file: %w", err)
	}

	return file, nil
}

func EnableUserSystemd(name string, log logging.Logger) error {
	command := exec.Command("/usr/bin/systemctl", "--user", "daemon-reload")
	if err := command.Run(); err != nil {
		return fmt.Errorf("systemctl daemon-reload: %w", err)
	}

	log.Infof("systemctl enable %q", name)
	command = exec.Command("/usr/bin/systemctl", "--user", "enable", name)
	if err := command.Run(); err != nil {
		return fmt.Errorf("systemctl enable: %w", err)
	}

	log.Infof("systemctl restart %q", name)
	command = exec.Command("/usr/bin/systemctl", "--user", "--no-block", "restart", name)
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
