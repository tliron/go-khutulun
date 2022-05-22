package agent

import (
	"fmt"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/logging/sink"
	"github.com/tliron/kutil/util"
)

//
// Gossip
//

type Gossip struct {
	Address string
	Port    int // memberlist default is 7946

	members     *memberlist.Memberlist
	queue       *memberlist.TransmitLimitedQueue
	onMessage   OnMessageFunc
	broadcaster *Broadcaster
	meta        []byte
}

func NewGossip(address string, port int) *Gossip {
	return &Gossip{
		Address: address,
		Port:    port,
	}
}

func (self *Gossip) Start() error {
	var err error

	if self.Address, err = util.ToReachableIPAddress(self.Address); err != nil {
		return err
	}

	config := memberlist.DefaultLANConfig()
	config.BindAddr = self.Address
	config.BindPort = self.Port
	config.AdvertisePort = self.Port
	config.Delegate = self
	config.Events = sink.NewMemberlistEventLog(gossipLog)
	config.Logger = sink.NewMemberlistStandardLog([]string{"gossip"})

	gossipLog.Noticef("starting server on %s", util.JoinIPAddressPort(config.BindAddr, config.BindPort))
	if self.members, err = memberlist.Create(config); err == nil {
		self.queue = &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				return self.members.NumMembers()
			},
		}
		return nil
	} else {
		return err
	}
}

func (self *Gossip) LocalGossipAddress() string {
	return self.members.LocalNode().Address()
}

func (self *Gossip) Announce() error {
	if self.broadcaster == nil {
		return nil
	}

	return self.broadcaster.SendJSON(NewAddHostCommand(self.LocalGossipAddress()))
}

func (self *Gossip) Stop() error {
	if self.members != nil {
		err := self.members.Leave(time.Second * 5)
		logging.CallAndLogError(self.members.Shutdown, "shutdown", gossipLog)
		return err
	} else {
		return nil
	}
}

type HostInformation struct {
	Name        string `json:"name"`
	GRPCAddress string `json:"grpcAddress"`
}

func (self *Gossip) ListHosts() []*HostInformation {
	var hosts []*HostInformation
	for _, node := range self.members.Members() {
		hosts = append(hosts, &HostInformation{
			Name:        node.Name,
			GRPCAddress: util.BytesToString(node.Meta),
		})
	}
	return hosts
}

func (self *Gossip) GetHost(name string) *HostInformation {
	for _, node := range self.members.Members() {
		if node.Name == name {
			return &HostInformation{
				Name:        node.Name,
				GRPCAddress: util.BytesToString(node.Meta),
			}
		}
	}
	return nil
}

func (self *Gossip) AddHosts(gossipAddresses []string) error {
	_, err := self.members.Join(gossipAddresses)
	return err
}

func (self *Gossip) SendJSON(host string, message any) error {
	if code, err := format.EncodeJSON(message, ""); err == nil {
		return self.Send(host, util.StringToBytes(code))
	} else {
		return err
	}
}

func (self *Gossip) Send(host string, message []byte) error {
	if node, ok := self.GetMember(host); ok {
		gossipLog.Debugf("sending message to %s: %s", host, message)
		return self.members.SendReliable(node, message)
	} else {
		return fmt.Errorf("member not found: %s", host)
	}
}

func (self *Gossip) GetMember(host string) (*memberlist.Node, bool) {
	for _, member := range self.members.Members() {
		if member.Name == host {
			return member, true
		}
	}
	return nil, false
}

// memberlist.Delegate interface
func (self *Gossip) NodeMeta(limit int) []byte {
	// limit is often 512
	if length := len(self.meta); length <= limit {
		return self.meta
	} else {
		gossipLog.Warningf("meta is too long: %d > %d", length, limit)
		return nil
	}
}

// memberlist.Delegate interface
func (self *Gossip) NotifyMsg(bytes []byte) {
	self.onMessage(bytes, false)
}

// memberlist.Delegate interface
func (self *Gossip) GetBroadcasts(overhead int, limit int) [][]byte {
	if self.queue != nil {
		return self.queue.GetBroadcasts(overhead, limit)
	} else {
		return nil
	}
}

// memberlist.Delegate interface
func (self *Gossip) LocalState(join bool) []byte {
	return nil
}

// memberlist.Delegate interface
func (self *Gossip) MergeRemoteState(buf []byte, join bool) {
}
