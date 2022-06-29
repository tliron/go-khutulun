package agent

import (
	"os"

	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/ard"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
)

type OnMessageFunc func(bytes []byte, broadcast bool)

//
// Agent
//

type Agent struct {
	host       string
	state      *sdk.State
	urlContext *urlpkg.Context
	gossip     *Gossip
}

func NewAgent(stateRootDir string) (*Agent, error) {
	if host, err := os.Hostname(); err == nil {
		return &Agent{
			host:       host,
			state:      sdk.NewState(stateRootDir),
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
		go self.handleCommand(message, broadcast)
	} else {
		log.Errorf("%s", err.Error())
	}
}
