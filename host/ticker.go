package host

import (
	"time"

	"github.com/tliron/kutil/util"
)

//
// Ticker
//

type Ticker struct {
	ticker *time.Ticker
	stop   chan struct{}
	f      func()
	lock   util.RWLocker
}

func NewTicker(frequency time.Duration, f func()) *Ticker {
	return &Ticker{
		stop: make(chan struct{}),
		f:    f,
		lock: util.NewDefaultRWLocker(),
	}
}

func (self *Ticker) Start() {
	self.ticker = time.NewTicker(TICKER_FREQUENCY)
	self.f()
	for {
		select {
		case <-self.stop:
			log.Info("stopping ticker")
			return

		case <-self.ticker.C:
			log.Info("tick")
			self.lock.Lock()
			self.f()
			self.lock.Unlock()
		}
	}
}

func (self *Ticker) Stop() {
	self.ticker.Stop()
	self.stop <- struct{}{}
}
