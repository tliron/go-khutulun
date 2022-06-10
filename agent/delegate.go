package agent

import (
	"fmt"
	"os"

	delegatepkg "github.com/tliron/khutulun/delegate"
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	cloutpkg "github.com/tliron/puccini/clout"
)

//
// Delegate
//

type Delegate interface {
	Name() (string, string)
	Delegate() delegatepkg.Delegate
	Release() error
}

func (self *Agent) NewDelegate(namespace string, delegateName string) (Delegate, error) {
	return self.NewPluginDelegate(namespace, delegateName)
}

//
// PluginDelegate
//

type PluginDelegate struct {
	namespace string
	name      string
	delegate  delegatepkg.Delegate
	client    *delegatepkg.DelegatePluginClient
}

func (self *Agent) NewPluginDelegate(namespace string, delegateName string) (*PluginDelegate, error) {
	if lock, err := self.lockPackage(namespace, "delegate", delegateName, false); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", delegateLog)

		command := self.getPackageMainFile(namespace, "delegate", delegateName)
		self := PluginDelegate{
			name:   delegateName,
			client: delegatepkg.NewDelegatePluginClient(delegateName, command),
		}

		if self.delegate, err = self.client.Delegate(); err == nil {
			return &self, nil
		} else {
			self.Release()
			return nil, err
		}
	} else if os.IsNotExist(err) {
		return nil, fmt.Errorf("delegate not found: %s/%s", namespace, delegateName)
	} else {
		return nil, err
	}
}

// Delegate interface
func (self *PluginDelegate) Name() (string, string) {
	return self.namespace, self.name
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

//
// Delegates
//

type Delegates struct {
	agent     *Agent
	delegates map[Namespaced]Delegate
}

type Namespaced struct {
	Namespace string
	Name      string
}

func NewNamespaced(namespace string, name string) Namespaced {
	return Namespaced{
		Namespace: namespace,
		Name:      name,
	}
}

func (self *Agent) NewDelegates() *Delegates {
	return &Delegates{
		agent:     self,
		delegates: make(map[Namespaced]Delegate),
	}
}

func (self *Delegates) NewDelegate(namespace string, delegateName string) (Delegate, error) {
	if delegate, err := self.agent.NewDelegate(namespace, delegateName); err == nil {
		self.delegates[NewNamespaced(namespace, delegateName)] = delegate
		if namespace_, _ := delegate.Name(); namespace_ != namespace {
			self.delegates[NewNamespaced(namespace_, delegateName)] = delegate
		}
		return delegate, nil
	} else {
		return nil, err
	}
}

func (self *Delegates) Get(namespace string, delegateName string) (delegatepkg.Delegate, error) {
	if delegate, ok := self.delegates[NewNamespaced(namespace, delegateName)]; ok {
		return delegate.Delegate(), nil
	} else if delegate, err := self.NewDelegate(namespace, delegateName); err == nil {
		return delegate.Delegate(), nil
	} else if namespace != "common" {
		return self.Get("common", delegateName)
	} else {
		return nil, err
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

func (self *Delegates) Fill(namespace string, coercedClout *cloutpkg.Clout) {
	for _, vertex := range coercedClout.Vertexes {
		for _, edge := range vertex.EdgesOut {
			if types, ok := ard.NewNode(edge.Properties).Get("types").StringMap(); ok {
				if _, ok := types["cloud.puccini.khutulun::Connection"]; ok {
					if delegateName, ok := ard.NewNode(edge.Properties).Get("attributes").Get("delegate").String(); ok {
						if _, err := self.Get(namespace, delegateName); err != nil {
							delegateLog.Errorf("%s", err.Error())
						}
					}
				}
			}
		}

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
