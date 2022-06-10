package agent

import (
	"time"
)

//
// Ticker
//

type Ticker struct {
	frequency time.Duration
	f         func()

	stop   chan struct{}
	ticker *time.Ticker
}

func NewTicker(frequency time.Duration, f func()) *Ticker {
	return &Ticker{
		frequency: frequency,
		f:         f,
		stop:      make(chan struct{}),
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
			self.f()
		}
	}
}

func (self *Ticker) Stop() {
	if self.ticker != nil {
		self.ticker.Stop()
	}
	self.stop <- struct{}{}
}
