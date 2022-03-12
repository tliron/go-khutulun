package conductor

import (
	"time"

	"github.com/tliron/kutil/logging"
)

//
// Ticker
//

type Ticker struct {
	ticker *time.Ticker
	stop   chan struct{}
	f      func()
	log    logging.Logger
}

func NewTicker(frequency time.Duration, f func(), log logging.Logger) *Ticker {
	return &Ticker{
		stop: make(chan struct{}),
		f:    f,
		log:  log,
	}
}

func (self *Ticker) Start() {
	self.ticker = time.NewTicker(FREQUENCY)
	self.f()
	for {
		select {
		case <-self.stop:
			self.log.Info("stopping ticker")
			return

		case <-self.ticker.C:
			self.log.Info("tick")
			self.f()
		}
	}
}

func (self *Ticker) Stop() {
	self.ticker.Stop()
	self.stop <- struct{}{}
}
