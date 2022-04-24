package conductor

import (
	"fmt"
	"net"
	"time"
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
	BroadcastAddress   string
	BroadcastPort      int

	conductor *Conductor
	watcher   *Watcher
	grpc      *GRPC
	http      *HTTP
	ticker    *Ticker
}

func NewServer(conductor *Conductor) *Server {
	return &Server{
		conductor: conductor,
	}
}

func (self *Server) Start(watcher bool, ticker bool) error {
	var err error

	if watcher {
		if self.watcher, err = NewWatcher(self.conductor, func(change Change, identifier []string) {
			if change != Changed {
				fmt.Printf("%s %v\n", change.String(), identifier)
			}
		}); err == nil {
			self.watcher.Start()
		} else {
			return err
		}
	}

	if self.GossipPort != 0 {
		self.conductor.cluster = NewCluster(self.GossipAddress, self.GossipPort, self.BroadcastProtocol, self.BroadcastInterface, self.BroadcastAddress, self.BroadcastPort)
		self.conductor.cluster.onMessage = self.conductor.onMessage
		if err := self.conductor.cluster.Start(); err != nil {
			return err
		}
	}

	if self.GRPCPort != 0 {
		self.grpc = NewGRPC(self.conductor, self.GRPCProtocol, self.GRPCAddress, self.GRPCPort)
		if err := self.grpc.Start(); err != nil {
			if self.conductor.cluster != nil {
				if err := self.conductor.cluster.Stop(); err != nil {
					log.Errorf("%s", err.Error())
				}
			}
			return err
		}
	}

	if self.HTTPPort != 0 {
		var err error
		if self.http, err = NewHTTP(self.conductor, self.HTTPProtocol, self.HTTPAddress, self.HTTPPort); err == nil {
			if err := self.http.Start(); err != nil {
				if self.grpc != nil {
					self.grpc.Stop()
				}
				if self.conductor.cluster != nil {
					if err := self.conductor.cluster.Stop(); err != nil {
						log.Errorf("%s", err.Error())
					}
				}
				return err
			}
		} else {
			return err
		}
	}

	if ticker {
		self.ticker = NewTicker(TICKER_FREQUENCY, func() {
			//self.conductor.Schedule()
			//self.conductor.Reconcile()
			if self.conductor.cluster != nil {
				if err := self.conductor.cluster.Announce(); err != nil {
					log.Errorf("%s", err.Error())
				}
			}
		})
		self.ticker.Start()
	}

	return nil
}

func (self *Server) Stop() error {
	var err error

	if self.ticker != nil {
		self.ticker.Stop()
	}

	if self.http != nil {
		if err_ := self.http.Stop(); err_ != nil {
			err = err_
		}
	}

	if self.grpc != nil {
		self.grpc.Stop()
	}

	if self.conductor.cluster != nil {
		if err_ := self.conductor.cluster.Stop(); err_ != nil {
			err = err_
		}
	}

	if self.watcher != nil {
		if err_ := self.watcher.Stop(); err_ != nil {
			err = err_
		}
	}

	return err
}
