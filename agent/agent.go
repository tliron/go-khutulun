package agent

import (
	contextpkg "context"
	"os"

	"github.com/tliron/exturl"
	"github.com/tliron/go-ard"
	"github.com/tliron/go-khutulun/sdk"
)

type OnMessageFunc func(bytes []byte, broadcast bool)

//
// Agent
//

type Agent struct {
	host       string
	state      *sdk.State
	urlContext *exturl.Context
	gossip     *Gossip
}

func NewAgent(stateRootDir string) (*Agent, error) {
	if host, err := os.Hostname(); err == nil {
		return &Agent{
			host:       host,
			state:      sdk.NewState(stateRootDir),
			urlContext: exturl.NewContext(),
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
	if message, err := ard.DecodeJSON(bytes, true); err == nil {
		go self.handleCommand(contextpkg.TODO(), message, broadcast)
	} else {
		log.Error(err.Error())
	}
}
