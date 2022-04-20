package conductor

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/tliron/kutil/logging/sink"
	"github.com/tliron/kutil/util"
)

const ADD_HOST = "khutulun.addHost:"

//
// Cluster
//

type Cluster struct {
	clusterPort      int
	broadcastNetwork string
	broadcastAddress string
	broadcastPort    int

	cluster      *memberlist.Memberlist
	clusterQueue *memberlist.TransmitLimitedQueue
	broadcaster  *Broadcaster
	receiver     *Receiver
}

func NewCluster() (*Cluster, error) {
	self := Cluster{
		clusterPort:      7946,
		broadcastNetwork: "udp4",
		broadcastAddress: "239.0.0.0",
		broadcastPort:    7947,
	}
	var err error
	broadcastAddress := fmt.Sprintf("%s:%d", self.broadcastAddress, self.broadcastPort)
	if self.broadcaster, err = NewBroadcaster(self.broadcastNetwork, broadcastAddress); err != nil {
		return nil, err
	}
	if self.receiver, err = NewReceiver(self.broadcastNetwork, broadcastAddress, self.receive); err != nil {
		return nil, err
	}
	return &self, nil
}

func (self *Cluster) Start() error {
	config := memberlist.DefaultLocalConfig()
	config.BindPort = self.clusterPort
	config.AdvertisePort = self.clusterPort
	config.Delegate = self
	config.Events = sink.NewMemberlistEventLog(clusterLog)
	//config.Logger =

	clusterLog.Notice("starting memberlist")
	var err error
	if self.cluster, err = memberlist.Create(config); err == nil {
		self.clusterQueue = &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				return self.cluster.NumMembers()
			},
		}

		if err := self.broadcaster.Start(); err != nil {
			return self.Stop()
		}

		self.receiver.Ignore = append(self.receiver.Ignore, self.broadcaster.Address())
		if err := self.receiver.Start(); err != nil {
			return self.Stop()
		}

		return nil
	} else {
		return err
	}
}

func (self *Cluster) Announce() error {
	address := self.cluster.LocalNode().Address()
	return self.broadcaster.SendString(ADD_HOST + address)
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
			address: fmt.Sprintf("%s:%d", node.Addr.String(), node.Port),
		})
	}
	return hosts
}

func (self *Cluster) AddHosts(hosts []string) error {
	_, err := self.cluster.Join(hosts)
	return err
}

// memberlist.Delegate interface
func (self *Cluster) NodeMeta(limit int) []byte {
	return nil
}

// memberlist.Delegate interface
func (self *Cluster) NotifyMsg(bytes []byte) {
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
	message_ := util.BytesToString(message)
	if strings.HasPrefix(message_, ADD_HOST) {
		host := message_[len(ADD_HOST):]
		if err := self.AddHosts([]string{host}); err != nil {
			clusterLog.Errorf("%s", err.Error())
		}
	} else {
		clusterLog.Errorf("received unsupported broadcast: %s", message_)
	}
}
