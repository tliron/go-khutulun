package agent

import (
	"fmt"
	"os"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-ard"
	delegatepkg "github.com/tliron/khutulun/delegate"
	cloutpkg "github.com/tliron/puccini/clout"
	cloututil "github.com/tliron/puccini/clout/util"
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

	delegate delegatepkg.Delegate
	client   *delegatepkg.DelegatePluginClient
	lock     fslock.Handle
}

func (self *Agent) NewPluginDelegate(namespace string, delegateName string) (*PluginDelegate, error) {
	if lock, err := self.state.LockPackage(namespace, "delegate", delegateName, false); err == nil {
		command := self.state.GetPackageMainFile(namespace, "delegate", delegateName)
		pluginDelegate := PluginDelegate{
			name:   delegateName,
			client: delegatepkg.NewDelegatePluginClient(delegateName, command),
			lock:   lock,
		}

		if pluginDelegate.delegate, err = pluginDelegate.client.Delegate(); err == nil {
			return &pluginDelegate, nil
		} else {
			commonlog.CallAndLogError(pluginDelegate.Release, "release", delegateLog)
			commonlog.CallAndLogError(lock.Unlock, "unlock", delegateLog)
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
	return self.lock.Unlock()
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
	for _, vertex := range cloututil.GetToscaNodeTemplates(coercedClout, "") {
		for _, relationship := range cloututil.GetToscaRelationships(vertex, "cloud.puccini.khutulun::Connection") {
			if delegateName, ok := ard.NewNode(relationship).Get("attributes", "delegate").String(); ok {
				if _, err := self.Get(namespace, delegateName); err != nil {
					delegateLog.Errorf("%s", err.Error())
				}
			}
		}

		for _, capability := range cloututil.GetToscaCapabilities(vertex, "cloud.puccini.khutulun::Activity") {
			if delegateName, ok := ard.NewNode(capability).Get("attributes", "delegate").String(); ok {
				if _, err := self.Get(namespace, delegateName); err != nil {
					delegateLog.Errorf("%s", err.Error())
				}
			}
		}
	}
}

func (self *Delegates) Release() {
	for _, delegate := range self.delegates {
		commonlog.CallAndLogError(delegate.Release, "release", delegateLog)
	}
}
