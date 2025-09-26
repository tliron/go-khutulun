package main

import (
	"fmt"

	"github.com/tliron/go-khutulun/delegate"
	cloutpkg "github.com/tliron/go-puccini/clout"
)

//
// Delegate
//

type Delegate struct {
	host string
}

// delegate.Delegate interface
func (self *Delegate) ProcessService(namespace string, serviceName string, phase string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	switch phase {
	case "schedule":
		return self.Schedule(namespace, serviceName, clout, coercedClout)

	case "reconcile":
		return self.Reconcile(namespace, serviceName, clout, coercedClout)

	default:
		return nil, nil, fmt.Errorf("unsupported phase: %s", phase)
	}
}
