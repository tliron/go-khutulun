package agent

import (
	contextpkg "context"

	"github.com/tliron/go-ard"
)

const (
	ADD_HOST        = "khutulun.addHost"
	PROCESS_SERVICE = "khutulun.processService"
)

func NewAddHostCommand(address string) map[string]any {
	command := make(map[string]any)
	command["command"] = ADD_HOST
	command["address"] = address
	return command
}

func NewProcessServiceCommand(namespace string, serviceName string, phase string) map[string]any {
	command := make(map[string]any)
	command["command"] = PROCESS_SERVICE
	command["namespace"] = namespace
	command["serviceName"] = serviceName
	command["phase"] = phase
	return command
}

func (self *Agent) handleCommand(context contextpkg.Context, message any, broadcast bool) {
	message_ := ard.With(message)
	command, _ := message_.Get("command").String()

	switch command {
	case ADD_HOST:
		address, _ := message_.Get("address").String()
		log.Infof("received addHost(%q)", address)
		if err := self.gossip.AddHosts([]string{address}); err != nil {
			log.Error(err.Error())
		}

	case PROCESS_SERVICE:
		namespace, _ := message_.Get("namespace").String()
		serviceName, _ := message_.Get("serviceName").String()
		phase, _ := message_.Get("phase").String()
		log.Infof("received processService(%q, %q, %q)", namespace, serviceName, phase)
		delegates := self.NewDelegates()
		defer delegates.Release()
		self.ProcessService(context, namespace, serviceName, phase, delegates)

	default:
		log.Errorf("received unsupported message: %s", message)
	}
}
