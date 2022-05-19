package agent

import (
	"os"

	"github.com/tliron/kutil/ard"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
)

type OnMessageFunc func(bytes []byte, broadcast bool)

const (
	ADD_HOST        = "khutulun.addHost"
	PROCESS_SERVICE = "khutulun.processService"
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

	switch command {
	case ADD_HOST:
		address, _ := ard.NewNode(message).Get("address").String()
		log.Infof("received addHost(%q)", address)
		if err := self.gossip.AddHosts([]string{address}); err != nil {
			log.Errorf("%s", err.Error())
		}

	case PROCESS_SERVICE:
		namespace, _ := ard.NewNode(message).Get("namespace").String()
		serviceName, _ := ard.NewNode(message).Get("serviceName").String()
		phase, _ := ard.NewNode(message).Get("phase").String()
		log.Infof("received processService(%q,%q,%q)", namespace, serviceName, phase)
		self.ProcessService(namespace, serviceName, phase)

	default:
		log.Errorf("received unsupported message: %s", message)
	}
}
