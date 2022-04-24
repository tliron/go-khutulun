package conductor

import (
	"os"

	"github.com/tliron/kutil/ard"
	urlpkg "github.com/tliron/kutil/url"
)

const (
	ADD_HOST           = "khutulun.addHost"
	RECONCILE_SERVICES = "khutulun.reconcileServices"
)

//
// Conductor
//

type Conductor struct {
	host       string
	statePath  string
	urlContext *urlpkg.Context
	cluster    *Cluster
}

func NewConductor(statePath string) (*Conductor, error) {
	if host, err := os.Hostname(); err == nil {
		return &Conductor{
			host:       host,
			statePath:  statePath,
			urlContext: urlpkg.NewContext(),
		}, nil
	} else {
		return nil, err
	}
}

func (self *Conductor) Release() error {
	return self.urlContext.Release()
}

// OnMessageFunc signature
func (self *Conductor) onMessage(message any, broadcast bool) {
	command, _ := ard.NewNode(message).Get("command").String()
	log.Infof("onMessage: %t, %s", broadcast, command)
	switch command {
	case ADD_HOST:
		address, _ := ard.NewNode(message).Get("address").String()
		if err := self.cluster.AddHosts([]string{address}); err != nil {
			log.Errorf("%s", err.Error())
		}

	case RECONCILE_SERVICES:
		identifiers, _ := ard.NewNode(message).Get("identifiers").List()
		for _, identifier := range identifiers {
			namespace, _ := ard.NewNode(identifier).Get("namespace").String()
			name, _ := ard.NewNode(identifier).Get("name").String()
			self.ReconcileService(namespace, name)
		}

	default:
		log.Errorf("received unsupported message: %s", message)
	}
}
