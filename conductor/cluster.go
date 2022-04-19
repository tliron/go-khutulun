package conductor

import (
	"fmt"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/tliron/kutil/logging/sink"
)

//
// Cluster
//

type Cluster struct {
	port         int
	cluster      *memberlist.Memberlist
	clusterQueue *memberlist.TransmitLimitedQueue
}

func NewCluster() *Cluster {
	return &Cluster{
		port: 7946,
	}
}

func (self *Cluster) Start() error {
	config := memberlist.DefaultLocalConfig()
	config.BindPort = self.port
	config.AdvertisePort = self.port
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
		return nil
	} else {
		return err
	}
}

func (self *Cluster) Stop() error {
	if self.cluster != nil {
		err := self.cluster.Leave(time.Second * 5)
		self.cluster.Shutdown()
		return err
	} else {
		return nil
	}
}

type Member struct {
	name    string
	address string
}

func (self *Cluster) ListMembers() []Member {
	var members []Member
	for _, node := range self.cluster.Members() {
		members = append(members, Member{
			name:    node.Name,
			address: fmt.Sprintf("%s:%d", node.Addr.String(), node.Port),
		})
	}
	return members
}

func (self *Cluster) AddMembers(hosts []string) error {
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
