package delegate

import (
	contextpkg "context"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/tliron/khutulun/api"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/logging/sink"
	"google.golang.org/grpc"
)

//
// DelegatePlugin
//

type DelegatePlugin struct {
	plugin.Plugin

	implementation Delegate // only for servers
}

// plugin.GRPCPlugin interface
func (self *DelegatePlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	api.RegisterDelegateServer(server, NewDelegateGRPCServer(self.implementation))
	return nil
}

// plugin.GRPCPlugin interface
func (p *DelegatePlugin) GRPCClient(context contextpkg.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	return NewDelegateGRPCClient(context, api.NewDelegateClient(client)), nil
}

//
// DelegatePluginClient
//

type DelegatePluginClient struct {
	client *plugin.Client
}

func NewDelegatePluginClient(name string, command string) *DelegatePluginClient {
	var config = plugin.ClientConfig{
		Plugins: map[string]plugin.Plugin{
			"delegate": new(DelegatePlugin),
		},
		Cmd:              exec.Command(command),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		HandshakeConfig:  handshakeConfig,
		Logger:           sink.NewHCLogger("khutulun.plugin."+name, nil),
	}

	return &DelegatePluginClient{
		client: plugin.NewClient(&config),
	}
}

func (self *DelegatePluginClient) Close() {
	self.client.Kill()
}

func (self *DelegatePluginClient) Delegate() (Delegate, error) {
	if protocol, err := self.client.Client(); err == nil {
		if service, err := protocol.Dispense("delegate"); err == nil {
			return service.(Delegate), nil
		} else {
			logging.CallAndLogError(protocol.Close, "close", log)
			return nil, err
		}
	} else {
		return nil, err
	}
}

//
// DelegatePluginServer
//

type DelegatePluginServer struct {
	plugin plugin.Plugin
}

func NewDelegatePluginServer(implementation Delegate) *DelegatePluginServer {
	return &DelegatePluginServer{
		plugin: &DelegatePlugin{implementation: implementation},
	}
}

func (self *DelegatePluginServer) Start() {
	var config = plugin.ServeConfig{
		Plugins: map[string]plugin.Plugin{
			"delegate": self.plugin,
		},
		GRPCServer:      plugin.DefaultGRPCServer,
		HandshakeConfig: handshakeConfig,
		Logger:          sink.NewHCLogger("khutulun.plugin.server", nil),
	}

	plugin.Serve(&config)
}
