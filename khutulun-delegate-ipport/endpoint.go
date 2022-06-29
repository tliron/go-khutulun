package main

import (
	"github.com/tliron/kutil/util"
)

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
