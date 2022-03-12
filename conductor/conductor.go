package conductor

import (
	"sync"

	urlpkg "github.com/tliron/kutil/url"
)

//
// Conductor
//

type Conductor struct {
	statePath  string
	urlContext *urlpkg.Context
	lock       sync.Mutex
}

func NewConductor(statePath string) *Conductor {
	return &Conductor{
		statePath:  statePath,
		urlContext: urlpkg.NewContext(),
	}
}

func (self *Conductor) Release() error {
	return self.urlContext.Release()
}
