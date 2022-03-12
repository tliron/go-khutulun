package configuration

import (
	"fmt"
	"os"
	userpkg "os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//
// Client
//

type Client struct {
	Clusters       map[string]Cluster `yaml:"clusters"`
	DefaultCluster string             `yaml:"default-cluster"`
}

type Cluster struct {
	IP   string `yaml:"ip"`
	Port int    `yaml:"port"`
}

func NewClient() *Client {
	return &Client{
		Clusters: make(map[string]Cluster),
	}
}

func LoadClient(path string) (*Client, error) {
	if path == "" {
		var err error
		if path, err = GetDefaultClientPath(); err != nil {
			return nil, err
		}
	}

	if file, err := os.Open(path); err == nil {
		decoder := yaml.NewDecoder(file)
		var client Client
		if err := decoder.Decode(&client); err == nil {
			if err := client.Validate(); err == nil {
				return &client, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func LoadOrNewClient(path string) (*Client, error) {
	client, err := LoadClient(path)
	if os.IsNotExist(err) {
		return NewClient(), nil
	} else {
		return client, err
	}
}

func (self *Client) GetCluster(name string) *Cluster {
	if name == "" {
		name = self.DefaultCluster
	}

	if name == "" {
		return nil
	}

	if self.Clusters != nil {
		if cluster, ok := self.Clusters[name]; ok {
			return &cluster
		}
	}

	return nil
}

func (self *Client) Validate() error {
	if self.DefaultCluster != "" {
		found := false
		if self.Clusters != nil {
			for name := range self.Clusters {
				if name == self.DefaultCluster {
					found = true
					break
				}
			}
		}
		if !found {
			return fmt.Errorf("default-cluster %q not found in clusters", self.DefaultCluster)
		}
	}

	return nil
}

func (self *Client) Save(path string) error {
	if path == "" {
		var err error
		if path, err = GetDefaultClientPath(); err != nil {
			return err
		}
	}

	if file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err == nil {
		encoder := yaml.NewEncoder(file)
		encoder.SetIndent(2)
		return encoder.Encode(self)
	} else {
		return err
	}
}

// Utils

func GetDefaultClientPath() (string, error) {
	if user, err := userpkg.Current(); err == nil {
		return filepath.Join(user.HomeDir, ".khutulun", "client.yaml"), nil
	} else {
		return "", err
	}
}
