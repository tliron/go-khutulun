package host

import (
	"fmt"
	"net"
	"time"

	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
)

const TICKER_FREQUENCY = 30 * time.Second

//
// Server
//

type Server struct {
	GRPCProtocol       string
	GRPCAddress        string
	GRPCPort           int
	HTTPProtocol       string
	HTTPAddress        string
	HTTPPort           int
	GossipAddress      string
	GossipPort         int
	BroadcastProtocol  string
	BroadcastInterface *net.Interface
	BroadcastAddress   string // https://en.wikipedia.org/wiki/Multicast_address
	BroadcastPort      int

	host        *Host
	grpc        *GRPC
	http        *HTTP
	gossip      *Gossip
	broadcaster *Broadcaster
	receiver    *Receiver
	watcher     *Watcher
	ticker      *Ticker
}

func NewServer(host *Host) *Server {
	return &Server{
		host: host,
	}
}

func (self *Server) Start(watcher bool, ticker bool) error {
	var err error

	if watcher {
		if self.watcher, err = NewWatcher(self.host, func(change Change, identifier []string) {
			if change != Changed {
				fmt.Printf("%s %v\n", change.String(), identifier)
			}
		}); err == nil {
			self.watcher.Start()
		} else {
			self.Stop()
			return err
		}
	}

	if self.GRPCPort != 0 {
		self.grpc = NewGRPC(self.host, self.GRPCProtocol, self.GRPCAddress, self.GRPCPort)
		if err := self.grpc.Start(); err != nil {
			self.Stop()
			return err
		}
	}

	if self.HTTPPort != 0 {
		var err error
		if self.http, err = NewHTTP(self.host, self.HTTPProtocol, self.HTTPAddress, self.HTTPPort); err == nil {
			if err := self.http.Start(); err != nil {
				self.Stop()
				return err
			}
		} else {
			self.Stop()
			return err
		}
	}

	if self.GossipPort != 0 {
		self.gossip = NewGossip(self.GossipAddress, self.GossipPort)
		self.gossip.onMessage = self.host.onMessage
		if self.grpc != nil {
			self.gossip.meta = util.StringToBytes(fmt.Sprintf("[%s]:%d", self.grpc.Address, self.grpc.Port))
		}
		if err := self.gossip.Start(); err != nil {
			self.Stop()
			return err
		}
		self.host.gossip = self.gossip
	}

	if self.BroadcastPort != 0 {
		if self.broadcaster, err = NewBroadcaster(self.BroadcastProtocol, self.BroadcastAddress, self.BroadcastPort); err != nil {
			self.Stop()
			return err
		}
		if self.gossip != nil {
			self.gossip.broadcaster = self.broadcaster
		}

		if self.receiver, err = NewReceiver(self.BroadcastProtocol, self.BroadcastInterface, self.BroadcastAddress, self.BroadcastPort, func(address *net.UDPAddr, message []byte) {
			self.host.onMessage(message, true)
		}); err != nil {
			self.Stop()
			return err
		}

		if err := self.broadcaster.Start(); err != nil {
			self.Stop()
			return err
		}

		self.receiver.Ignore = append(self.receiver.Ignore, self.broadcaster.Address())
		if err := self.receiver.Start(); err != nil {
			self.Stop()
			return err
		}
	}

	if ticker {
		self.ticker = NewTicker(TICKER_FREQUENCY, func() {
			//self.host.Schedule()
			//self.host.Reconcile()
			if self.gossip != nil {
				logging.CallAndLogError(self.gossip.Announce, "announce", log)
			}
		})
		self.ticker.Start()
	}

	return nil
}

func (self *Server) Stop() {
	if self.ticker != nil {
		self.ticker.Stop()
	}

	if self.receiver != nil {
		self.receiver.Stop()
	}

	if self.broadcaster != nil {
		self.broadcaster.Stop()
	}

	if self.gossip != nil {
		logging.CallAndLogError(self.gossip.Stop, "stop", gossipLog)
	}

	if self.http != nil {
		logging.CallAndLogError(self.http.Stop, "stop", httpLog)
	}

	if self.grpc != nil {
		self.grpc.Stop()
	}

	if self.watcher != nil {
		logging.CallAndLogError(self.watcher.Stop, "stop", watcherLog)
	}
}
