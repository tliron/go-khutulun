package agent

import (
	"fmt"
	"net"
	"time"

	"github.com/tliron/commonlog"
	"github.com/tliron/go-khutulun/sdk"
	"github.com/tliron/go-kutil/util"
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

	agent       *Agent
	grpc        *GRPC
	http        *HTTP
	gossip      *Gossip
	broadcaster *Broadcaster
	receiver    *Receiver
	watcher     *sdk.Watcher
	ticker      *Ticker
}

func NewServer(agent *Agent) *Server {
	return &Server{
		agent: agent,
	}
}

func (self *Server) Start(watcher bool, ticker bool) error {
	var err error

	var host sdk.Host
	if host.Address, err = util.ToReachableIPAddress(self.GRPCAddress); err != nil {
		return err
	}
	if err := self.agent.state.SetHost(self.agent.host, &host); err != nil {
		return err
	}

	// TODO?
	watcher = false

	if watcher {
		if self.watcher, err = sdk.NewWatcher(self.agent.state, func(change sdk.Change, identifier []string) {
			if change != sdk.Changed {
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
		self.grpc = NewGRPC(self.agent, self.GRPCProtocol, self.GRPCAddress, self.GRPCPort)
		if err := self.grpc.Start(); err != nil {
			self.Stop()
			return err
		}
	}

	if self.HTTPPort != 0 {
		var err error
		if self.http, err = NewHTTP(self.agent, self.HTTPProtocol, self.HTTPAddress, self.HTTPPort); err == nil {
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
		self.gossip.onMessage = self.agent.onMessage
		if self.grpc != nil {
			self.gossip.meta = util.StringToBytes(util.JoinIPAddressPort(self.grpc.Address, self.grpc.Port))
		}
		if err := self.gossip.Start(); err != nil {
			self.Stop()
			return err
		}
		self.agent.gossip = self.gossip
	}

	if self.BroadcastPort != 0 {
		self.broadcaster = NewBroadcaster(self.BroadcastProtocol, self.BroadcastAddress, self.BroadcastPort)
		if self.gossip != nil {
			self.gossip.broadcaster = self.broadcaster
		}

		self.receiver = NewReceiver(self.BroadcastProtocol, self.BroadcastInterface, self.BroadcastAddress, self.BroadcastPort, func(address *net.UDPAddr, message []byte) {
			self.agent.onMessage(message, true)
		})

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
				commonlog.CallAndLogError(self.gossip.Announce, "announce", log)
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
		commonlog.CallAndLogError(self.gossip.Stop, "stop", gossipLog)
	}

	if self.http != nil {
		commonlog.CallAndLogError(self.http.Stop, "stop", httpLog)
	}

	if self.grpc != nil {
		self.grpc.Stop()
	}

	if self.watcher != nil {
		commonlog.CallAndLogError(self.watcher.Stop, "stop watcher", log)
	}
}
