package sdk

import (
	"github.com/tliron/commonlog"
	"gopkg.in/yaml.v3"
)

func (self *State) GetHost(name string) (*Host, error) {
	var host Host
	if reader, err := self.LockAndOpenPackageFile("common", "host", name, "host.yaml"); err == nil {
		defer commonlog.CallAndLogError(reader.Close, "close", stateLog)

		if err := yaml.NewDecoder(reader).Decode(&host); err == nil {
			return &host, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *State) SetHost(name string, host *Host) error {
	if writer, err := self.LockAndCreatePackageFile("common", "host", name, "host.yaml"); err == nil {
		defer commonlog.CallAndLogError(writer.Close, "close", stateLog)

		return yaml.NewEncoder(writer).Encode(host)
	} else {
		return err
	}
}

//
// Host
//

type Host struct {
	Address string `yaml:"address"`
}
