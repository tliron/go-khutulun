package main

import (
	"fmt"
	"strconv"

	"github.com/tliron/go-khutulun/delegate"
	"github.com/tliron/go-khutulun/sdk"
	"github.com/tliron/go-kutil/util"
	cloutpkg "github.com/tliron/go-puccini/clout"
)

// systemctl --machine user@.host --user
// https://superuser.com/a/1461905

func (self *Delegate) Reconcile(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	processes, err := sdk.GetCloutProcesses(coercedClout)
	if err != nil {
		return nil, nil, err
	}
	if len(processes) == 0 {
		return nil, nil, nil
	}

	for _, process := range processes {
		if process.Host == self.host {
			//format.WriteGo(container, logging.GetWriter(), " ")
			if err := self.CreateProcessUserService(process); err != nil {
				log.Errorf("instantiate: %s", err.Error())
			}
		}
	}

	return nil, nil, nil
}

func (self *Delegate) CreateProcessUserService(process *sdk.Process) error {
	serviceName := fmt.Sprintf("%s-%s.service", servicePrefix, process.Name)

	file, err := sdk.CreateUserSystemdFile(serviceName, log)
	if err != nil {
		return err
	}

	var command string
	if len(process.Arguments) > 0 {
		command = fmt.Sprintf("%q %s", process.Command, util.JoinQuote(process.Arguments, " "))
	} else {
		command = strconv.Quote(process.Command)
	}

	text := `
[Unit]
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
Restart=always
ExecStart=%s

[Install]
WantedBy=default.target
`
	text = fmt.Sprintf(text, command)

	_, err = file.WriteString(text)
	if err != nil {
		file.Close()
		return fmt.Errorf("write: %w", err)
	}

	file.Close()
	return sdk.EnableUserSystemd(serviceName, log)
}
