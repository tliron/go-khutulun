package main

import (
	"github.com/tliron/khutulun/delegate"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) Reconcile(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	return nil, nil, nil
}
