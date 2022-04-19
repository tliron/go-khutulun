package conductor

import (
	urlpkg "github.com/tliron/kutil/url"
)

//
// Conductor
//

type Conductor struct {
	statePath  string
	urlContext *urlpkg.Context
	cluster    *Cluster
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
