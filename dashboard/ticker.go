package dashboard

import (
	"time"

	"github.com/rivo/tview"
)

//
// Ticker
//

type Ticker struct {
	application *tview.Application
	frequency   time.Duration
	f           func()

	stop   chan struct{}
	ticker *time.Ticker
}

func NewTicker(application *tview.Application, frequency time.Duration, f func()) *Ticker {
	return &Ticker{
		application: application,
		frequency:   frequency,
		f:           f,
		stop:        make(chan struct{}),
	}
}

func (self *Ticker) Start() {
	self.f()
	self.ticker = time.NewTicker(self.frequency)
	go func() {
		for {
			select {
			case <-self.stop:
				return

			case <-self.ticker.C:
				self.application.QueueUpdateDraw(self.f)
			}
		}
	}()
}

func (self *Ticker) Stop() {
	if self.ticker != nil {
		self.ticker.Stop()
	}
	self.stop <- struct{}{}
}
