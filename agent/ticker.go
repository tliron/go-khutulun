package agent

import (
	"time"

	"github.com/tliron/kutil/util"
)

//
// Ticker
//

type Ticker struct {
	frequency time.Duration
	f         func()

	stop   chan struct{}
	lock   util.RWLocker
	ticker *time.Ticker
}

func NewTicker(frequency time.Duration, f func()) *Ticker {
	return &Ticker{
		frequency: frequency,
		f:         f,
		stop:      make(chan struct{}),
		lock:      util.NewDefaultRWLocker(),
	}
}

func (self *Ticker) Start() {
	self.f()
	self.ticker = time.NewTicker(self.frequency)
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
	if self.ticker != nil {
		self.ticker.Stop()
	}
	self.stop <- struct{}{}
}
