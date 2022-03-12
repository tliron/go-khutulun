package plugin

import (
	"net/rpc"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/tliron/kutil/logging/sink"
)

//
// Runnable
//

type Runnable interface {
	Instantiate(config map[string]any) error
}

//
// RunnableRPCServer
//

type RunnableRPCServer struct {
	implementation Runnable
}

func NewRunnableRPCServer(implementation Runnable) *RunnableRPCServer {
	return &RunnableRPCServer{implementation: implementation}
}

// net/rpc signature
func (s *RunnableRPCServer) Instantiate(request *map[string]any, response *struct{}) error {
	return s.implementation.Instantiate(*request)
}

//
// RunnableRPCClient
//

type RunnableRPCClient struct {
	client *rpc.Client
}

func NewRunRPCClient(client *rpc.Client) *RunnableRPCClient {
	return &RunnableRPCClient{client}
}

// Runnable interface
func (self *RunnableRPCClient) Instantiate(config map[string]any) error {
	var r struct{}
	return self.client.Call("Plugin.Instantiate", &config, &r)
}

//
// RunnablePlugin
//

type RunnablePlugin struct {
	implementation Runnable
}

// plugin.Plugin interface
func (self *RunnablePlugin) Server(broker *plugin.MuxBroker) (any, error) {
	return NewRunnableRPCServer(self.implementation), nil
}

// plugin.Plugin interface
func (self *RunnablePlugin) Client(broker *plugin.MuxBroker, client *rpc.Client) (any, error) {
	return NewRunRPCClient(client), nil
}

//
// RunnableClient
//

type RunnableClient struct {
	client *plugin.Client
}

func NewRunnableClient(name string, command string) *RunnableClient {
	var config = plugin.ClientConfig{
		Plugins: map[string]plugin.Plugin{
			"run": new(RunnablePlugin),
		},
		Cmd:              exec.Command(command),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		HandshakeConfig:  handshakeConfig,
		Logger:           sink.NewHCLogger("khutulun.plugin."+name, nil),
	}

	return &RunnableClient{
		client: plugin.NewClient(&config),
	}
}

func (self *RunnableClient) Close() {
	self.client.Kill()
}

func (self *RunnableClient) Runnable() (Runnable, error) {
	if protocol, err := self.client.Client(); err == nil {
		if r, err := protocol.Dispense("run"); err == nil {
			return r.(Runnable), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

//
// RunnableServer
//

type RunnableServer struct {
	plugin *RunnablePlugin
}

func NewRunnableServer(implementation Runnable) *RunnableServer {
	return &RunnableServer{
		plugin: &RunnablePlugin{implementation: implementation},
	}
}

func (self *RunnableServer) Start() {
	var config = plugin.ServeConfig{
		Plugins: map[string]plugin.Plugin{
			"run": self.plugin,
		},
		HandshakeConfig: handshakeConfig,
		Logger:          sink.NewHCLogger("khutulun.plugin.server", nil),
	}

	plugin.Serve(&config)
}
