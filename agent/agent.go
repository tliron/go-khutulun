package agent

import (
	"os"

	"github.com/tliron/kutil/ard"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
)

type OnMessageFunc func(bytes []byte, broadcast bool)

const (
	ADD_HOST           = "khutulun.addHost"
	RECONCILE_SERVICES = "khutulun.reconcileServices"
)

//
// Agent
//

type Agent struct {
	host       string
	statePath  string
	urlContext *urlpkg.Context
	gossip     *Gossip
}

func NewAgent(statePath string) (*Agent, error) {
	if host, err := os.Hostname(); err == nil {
		return &Agent{
			host:       host,
			statePath:  statePath,
			urlContext: urlpkg.NewContext(),
		}, nil
	} else {
		return nil, err
	}
}

func (self *Agent) Release() error {
	return self.urlContext.Release()
}

// OnMessageFunc signature
func (self *Agent) onMessage(bytes []byte, broadcast bool) {
	if message, _, err := ard.DecodeJSON(util.BytesToString(bytes), false); err == nil {
		go self.handleMessage(message, broadcast)
	} else {
		log.Errorf("%s", err.Error())
	}
}

func (self *Agent) handleMessage(message any, broadcast bool) {
	command, _ := ard.NewNode(message).Get("command").String()

	if broadcast {
		log.Infof("received broadcast message: %s()", command)
	} else {
		log.Infof("received message: %s()", command)
	}

	switch command {
	case ADD_HOST:
		address, _ := ard.NewNode(message).Get("address").String()
		if err := self.gossip.AddHosts([]string{address}); err != nil {
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
