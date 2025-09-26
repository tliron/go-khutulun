package main

import (
	"os"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-khutulun/sdk"
	"github.com/tliron/go-kutil/util"
	"gopkg.in/yaml.v2"
)

func LockAndGetHostEndpoints(state *sdk.State, host string) (fslock.Handle, *HostEndpoints, error) {
	hostEndpoints := NewHostEndpoints(host, "")

	if lock, err := state.LockPackage("common", "host", host, false); err == nil {
		if reader, err := state.OpenPackageFile("common", "host", host, "endpoints.yaml"); err == nil {
			if err := yaml.NewDecoder(reader).Decode(hostEndpoints); err == nil {
				return lock, hostEndpoints, nil
			} else {
				commonlog.CallAndLogError(lock.Unlock, "unlock", log)
				return nil, nil, err
			}
		} else if os.IsNotExist(err) {
			var host_ sdk.Host
			if reader, err := state.OpenPackageFile("common", "host", host, "host.yaml"); err == nil {
				if err := yaml.NewDecoder(reader).Decode(host_); err == nil {
					hostEndpoints.Address = host_.Address
					hostEndpoints.AddPortRange(9000, 9999)
					return lock, hostEndpoints, nil
				} else {
					commonlog.CallAndLogError(lock.Unlock, "unlock", log)
					return nil, nil, err
				}
			} else {
				commonlog.CallAndLogError(lock.Unlock, "unlock", log)
				return nil, nil, err
			}
		} else {
			commonlog.CallAndLogError(lock.Unlock, "unlock", log)
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

func SetHostEndpoints(state *sdk.State, hostEndpoints *HostEndpoints) error {
	if writer, err := state.CreatePackageFile("common", "host", hostEndpoints.Host, "endpoints.yaml"); err == nil {
		defer commonlog.CallAndLogError(writer.Close, "close", log)

		return yaml.NewEncoder(writer).Encode(hostEndpoints)
	} else {
		return err
	}
}

//
// Endpoint
//

type Endpoint struct {
	Namespace string `yaml:"namespace"`
	Service   string `yaml:"service"`
	Name      string `yaml:"name"`

	Host    string `yaml:"host"`
	Address string `yaml:"address"`
	Port    int64  `yaml:"port"`
}

//
// PortRange
//

type PortRange struct {
	Start int64 `yaml:"start"`
	End   int64 `yaml:"end"`
}

//
// HostEndpoints
//

type HostEndpoints struct {
	Host         string      `yaml:"host"`
	Address      string      `yaml:"address"`
	Reservations []*Endpoint `yaml:"reservations"`
	PortRanges   []PortRange `yaml:"portRanges"`

	lock util.RWLocker
}

func NewHostEndpoints(host string, address string) *HostEndpoints {
	return &HostEndpoints{
		Host:    host,
		Address: address,
		lock:    util.NewDefaultRWLocker(),
	}
}

func (self *HostEndpoints) AddPortRange(start int64, end int64) {
	self.PortRanges = append(self.PortRanges, PortRange{start, end})
}

func (self *HostEndpoints) GetReservation(port int64) *Endpoint {
	self.lock.RLock()
	defer self.lock.RUnlock()

	return self.getReservation(port)
}

func (self *HostEndpoints) getReservation(port int64) *Endpoint {
	for _, endpoint := range self.Reservations {
		if port == endpoint.Port {
			return endpoint
		}
	}
	return nil
}

func (self *HostEndpoints) Reserve(endpoint *Endpoint) bool {
	self.lock.Lock()
	defer self.lock.Unlock()

	// Is already reserved?
	for _, endpoint_ := range self.Reservations {
		if (endpoint.Namespace == endpoint_.Namespace) && (endpoint.Service == endpoint_.Service) && (endpoint.Name == endpoint_.Name) {
			endpoint.Address = endpoint_.Address
			endpoint.Host = endpoint_.Host
			endpoint.Port = endpoint_.Port
			return true
		}
	}

	for _, portRange := range self.PortRanges {
		for port := portRange.Start; port <= portRange.End; port++ {
			if self.getReservation(port) == nil {
				endpoint.Address = self.Address
				endpoint.Host = self.Host
				endpoint.Port = port
				self.Reservations = append(self.Reservations, endpoint)
				return true
			}
		}
	}
	return false
}

func (self *HostEndpoints) Release(port int64) *Endpoint {
	self.lock.Lock()
	defer self.lock.Unlock()

	for index, endpoint := range self.Reservations {
		if port == endpoint.Port {
			self.Reservations = append(self.Reservations[:index], self.Reservations[index+1:]...)
			return endpoint
		}
	}
	return nil
}
