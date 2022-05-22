package agent

import (
	"fmt"
	"os"

	"github.com/danjacques/gofslock/fslock"
	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Agent) GetDelegateCommand(namespace string, delegateName string) (string, fslock.Handle, error) {
	if lock, err := self.lockPackage(namespace, "delegate", delegateName, false); err == nil {
		return self.getPackageMainFile(namespace, "delegate", delegateName), lock, nil
	} else if os.IsNotExist(err) {
		namespace = "common"
		if lock, err := self.lockPackage(namespace, "delegate", delegateName, false); err == nil {
			return self.getPackageMainFile(namespace, "delegate", delegateName), lock, nil
		} else if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("delegate not found: %s/%s", namespace, delegateName)
		} else {
			return "", nil, err
		}
	} else {
		return "", nil, err
	}
}

//
// Delegate
//

type Delegate interface {
	Delegate() delegatepkg.Delegate
	Release() error
}

//
// PluginDelegate
//

type PluginDelegate struct {
	delegate delegatepkg.Delegate
	client   *delegatepkg.DelegatePluginClient
}

func (self *Agent) NewPluginDelegate(namespace string, delegateName string) (*PluginDelegate, error) {
	if command, lock, err := self.GetDelegateCommand(namespace, delegateName); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", delegateLog)
		var self PluginDelegate
		self.client = delegatepkg.NewDelegatePluginClient(delegateName, command)
		if self.delegate, err = self.client.Delegate(); err == nil {
			return &self, nil
		} else {
			self.client.Close()
			return nil, err
		}
	} else {
		return nil, err
	}
}

// Delegate interface
func (self *PluginDelegate) Delegate() delegatepkg.Delegate {
	return self.delegate
}

// Delegate interface
func (self *PluginDelegate) Release() error {
	self.client.Close()
	return nil
}

func (self *Agent) GetDelegate(namespace string, delegateName string) (Delegate, error) {
	return self.NewPluginDelegate(namespace, delegateName)
}

//
// Delegates
//

type Delegates struct {
	agent     *Agent
	delegates map[string]Delegate
}

func (self *Agent) NewDelegates() *Delegates {
	return &Delegates{
		agent:     self,
		delegates: make(map[string]Delegate),
	}
}

func (self *Delegates) All() []delegatepkg.Delegate {
	delegates := make([]delegatepkg.Delegate, len(self.delegates))
	index := 0
	for _, delegate := range self.delegates {
		delegates[index] = delegate.Delegate()
		index++
	}
	return delegates
}

func (self *Delegates) Get(namespace string, delegateName string) (delegatepkg.Delegate, error) {
	if delegate, ok := self.delegates[delegateName]; ok {
		return delegate.Delegate(), nil
	} else if delegate, err := self.agent.GetDelegate(namespace, delegateName); err == nil {
		self.delegates[delegateName] = delegate
		return delegate.Delegate(), nil
	} else {
		return nil, err
	}
}

func (self *Delegates) Fill(namespace string, coercedClout *cloutpkg.Clout) {
	for _, vertex := range coercedClout.Vertexes {
		if capabilities, ok := ard.NewNode(vertex.Properties).Get("capabilities").StringMap(); ok {
			for _, capability := range capabilities {
				if types, ok := ard.NewNode(capability).Get("types").StringMap(); ok {
					if _, ok := types["cloud.puccini.khutulun::Runnable"]; ok {
						if delegateName, ok := ard.NewNode(capability).Get("attributes").Get("delegate").String(); ok {
							if _, err := self.Get(namespace, delegateName); err != nil {
								delegateLog.Errorf("%s", err.Error())
							}
						}
					}
				}
			}
		}
	}
}

func (self *Delegates) Release() {
	for _, delegate := range self.delegates {
		logging.CallAndLogError(delegate.Release, "release", delegateLog)
	}
}
