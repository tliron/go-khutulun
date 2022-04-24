package conductor

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/logging/sink"
	"github.com/tliron/kutil/util"
)

type OnMessageFunc func(message any, broadcast bool)

//
// Cluster
//

type Cluster struct {
	GossipAddress      string
	GossipPort         int // memberlist default is 7946
	BroadcastProtocol  string
	BroadcastInterface *net.Interface
	BroadcastAddress   string // https://en.wikipedia.org/wiki/Multicast_address
	BroadcastPort      int

	onMessage    OnMessageFunc
	cluster      *memberlist.Memberlist
	clusterQueue *memberlist.TransmitLimitedQueue
	broadcaster  *Broadcaster
	receiver     *Receiver
}

func NewCluster(gossipAddress string, gossipPort int, broadcastProtocol string, broadcastInterface *net.Interface, broadcastAddress string, broadcastPort int) *Cluster {
	return &Cluster{
		GossipAddress:      gossipAddress,
		GossipPort:         gossipPort,
		BroadcastProtocol:  broadcastProtocol,
		BroadcastInterface: broadcastInterface,
		BroadcastAddress:   broadcastAddress,
		BroadcastPort:      broadcastPort,
	}
}

func (self *Cluster) Start() error {
	var err error

	if self.GossipAddress, err = toReachableAddress(self.GossipAddress); err != nil {
		return err
	}

	if self.BroadcastPort != 0 {
		if self.broadcaster, err = NewBroadcaster(self.BroadcastProtocol, self.BroadcastAddress, self.BroadcastPort); err != nil {
			return err
		}
		if self.receiver, err = NewReceiver(self.BroadcastProtocol, self.BroadcastInterface, self.BroadcastAddress, self.BroadcastPort, self.receive); err != nil {
			return err
		}
	}

	config := memberlist.DefaultLANConfig()
	config.BindAddr = self.GossipAddress
	config.BindPort = self.GossipPort
	config.AdvertisePort = self.GossipPort
	config.Delegate = self
	config.Events = sink.NewMemberlistEventLog(clusterLog)
	config.Logger = sink.NewMemberlistStandardLog([]string{"khutulun", "memberlist"})

	clusterLog.Notice("starting memberlist")
	if self.cluster, err = memberlist.Create(config); err == nil {
		self.clusterQueue = &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				return self.cluster.NumMembers()
			},
		}

		if err := self.broadcaster.Start(); err != nil {
			clusterLog.Errorf("%s", err.Error())
			return self.Stop()
		}

		self.receiver.Ignore = append(self.receiver.Ignore, self.broadcaster.Address())
		if err := self.receiver.Start(); err != nil {
			clusterLog.Errorf("%s", err.Error())
			return self.Stop()
		}

		return nil
	} else {
		return err
	}
}

func (self *Cluster) LocalGossipAddress() string {
	//return self.reachableGossipAddress
	return self.cluster.LocalNode().Address()
}

func (self *Cluster) Announce() error {
	command := make(map[string]any)
	command["command"] = ADD_HOST
	command["address"] = self.LocalGossipAddress()
	return self.broadcaster.SendJSON(command)
}

func (self *Cluster) Stop() error {
	if self.receiver != nil {
		self.receiver.Stop()
	}

	if self.broadcaster != nil {
		self.broadcaster.Stop()
	}

	if self.cluster != nil {
		err := self.cluster.Leave(time.Second * 5)
		self.cluster.Shutdown()
		return err
	} else {
		return nil
	}
}

type Host struct {
	name    string
	address string
}

func (self *Cluster) ListHosts() []Host {
	var hosts []Host
	for _, node := range self.cluster.Members() {
		hosts = append(hosts, Host{
			name:    node.Name,
			address: fmt.Sprintf("[%s]:%d", node.Addr.String(), node.Port),
		})
	}
	return hosts
}

func (self *Cluster) AddHosts(hosts []string) error {
	_, err := self.cluster.Join(hosts)
	return err
}

func (self *Cluster) SendJSON(host string, message any) (bool, error) {
	if code, err := format.EncodeJSON(message, ""); err == nil {
		return self.Send(host, util.StringToBytes(code))
	} else {
		return false, err
	}
}

func (self *Cluster) Send(host string, message []byte) (bool, error) {
	if node, ok := self.GetMember(host); ok {
		clusterLog.Infof("sending message to %s: %s", host, message)
		return true, self.cluster.SendReliable(node, message)
	} else {
		return false, nil
	}
}

func (self *Cluster) GetMember(host string) (*memberlist.Node, bool) {
	for _, member := range self.cluster.Members() {
		if member.Name == host {
			return member, true
		}
	}
	return nil, false
}

// memberlist.Delegate interface
func (self *Cluster) NodeMeta(limit int) []byte {
	return nil
}

// memberlist.Delegate interface
func (self *Cluster) NotifyMsg(bytes []byte) {
	if message, _, err := ard.DecodeJSON(util.BytesToString(bytes), false); err == nil {
		go self.onMessage(message, false)
	} else {
		clusterLog.Errorf("%s", err.Error())
	}
}

// memberlist.Delegate interface
func (self *Cluster) GetBroadcasts(overhead int, limit int) [][]byte {
	return self.clusterQueue.GetBroadcasts(overhead, limit)
}

// memberlist.Delegate interface
func (self *Cluster) LocalState(join bool) []byte {
	return nil
}

// memberlist.Delegate interface
func (self *Cluster) MergeRemoteState(buf []byte, join bool) {
}

func (self *Cluster) receive(address *net.UDPAddr, message []byte) {
	if message_, _, err := ard.DecodeJSON(util.BytesToString(message), false); err == nil {
		go self.onMessage(message_, true)
	} else {
		clusterLog.Errorf("%s", err.Error())
	}
}
