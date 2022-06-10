package agent

import (
	"github.com/tliron/kutil/ard"
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

func (self *Agent) handleCommand(message any, broadcast bool) {
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
		log.Infof("received processService(%q, %q, %q)", namespace, serviceName, phase)
		delegates := self.NewDelegates()
		defer delegates.Release()
		self.ProcessService(namespace, serviceName, phase, delegates)

	default:
		log.Errorf("received unsupported message: %s", message)
	}
}
