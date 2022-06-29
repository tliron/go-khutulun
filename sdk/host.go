package sdk

import (
	"github.com/tliron/kutil/logging"
	"gopkg.in/yaml.v3"
)

func (self *State) GetHost(name string) (*Host, error) {
	var host Host
	if err := self.GetHostFileYAML(name, "host.yaml", &host); err == nil {
		return &host, nil
	} else {
		return nil, err
	}
}

func (self *State) SetHost(name string, host *Host) error {
	return self.SetHostFileYAML(name, "host.yaml", host)
}

func (self *State) GetHostFileYAML(name string, path string, contentPointer any) error {
	if reader, err := self.OpenPackageFile("common", "host", name, path); err == nil {
		defer logging.CallAndLogError(reader.Close, "close", stateLog)

		return yaml.NewDecoder(reader).Decode(contentPointer)
	} else {
		return err
	}
}

func (self *State) SetHostFileYAML(name string, path string, content any) error {
	if writer, err := self.CreatePackageFile("common", "host", name, path); err == nil {
		defer logging.CallAndLogError(writer.Close, "close", stateLog)

		return yaml.NewEncoder(writer).Encode(content)
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
