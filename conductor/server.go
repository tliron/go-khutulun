package conductor

import (
	"time"
)

const TICKER_FREQUENCY = 10 * time.Second

//
// Server
//

type Server struct {
	conductor *Conductor
	grpc      *GRPC
	http      *HTTP
	ticker    *Ticker
}

func NewServer(conductor *Conductor) *Server {
	return &Server{conductor: conductor}
}

func (self *Server) Start(cluster bool, grpc bool, http bool, reconcile bool) error {
	if cluster {
		var err error
		if self.conductor.cluster, err = NewCluster(); err == nil {
			if err := self.conductor.cluster.Start(); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if grpc {
		self.grpc = NewGRPC(self.conductor)
		if err := self.grpc.Start(); err != nil {
			if self.conductor.cluster != nil {
				self.conductor.cluster.Stop()
			}
			return err
		}
	}

	if http {
		var err error
		if self.http, err = NewHTTP(self.conductor); err == nil {
			if err := self.http.Start(); err != nil {
				if self.grpc != nil {
					self.grpc.Stop()
				}
				if self.conductor.cluster != nil {
					self.conductor.cluster.Stop()
				}
				return err
			}
		} else {
			return err
		}
	}

	if reconcile {
		self.ticker = NewTicker(TICKER_FREQUENCY, func() {
			self.conductor.Schedule()
			self.conductor.Reconcile()
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

	return err
}
