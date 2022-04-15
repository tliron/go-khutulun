package plugin

import (
	contextpkg "context"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/util"
	"github.com/tliron/kutil/logging/sink"
	"github.com/tliron/kutil/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//
// Runnable
//

type Runnable interface {
	Instantiate(config any) error
	Interact(server util.Interactor, first *api.Interaction) error
}

//
// RunnableGRPCServer
//

type RunnableGRPCServer struct {
	api.UnimplementedPluginServer

	implementation Runnable
}

func NewRunnableGRPCServer(implementation Runnable) *RunnableGRPCServer {
	return &RunnableGRPCServer{implementation: implementation}
}

// api.PluginServer interface
func (self *RunnableGRPCServer) Instantiate(context contextpkg.Context, config *api.Config) (*emptypb.Empty, error) {
	if err := self.implementation.Instantiate(config.Config.AsMap()); err == nil {
		return new(emptypb.Empty), nil
	} else {
		return new(emptypb.Empty), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.PluginServer interface
func (self *RunnableGRPCServer) Interact(server api.Plugin_InteractServer) error {
	return util.Interact(server, map[string]util.InteractFunc{
		"runnable": func(first *api.Interaction) error {
			return self.implementation.Interact(server, first)
		},
	})
}

//
// RunnableGRPCClient
//

type RunnableGRPCClient struct {
	context contextpkg.Context
	client  api.PluginClient
}

func NewRunnableGRPCClient(context contextpkg.Context, client api.PluginClient) *RunnableGRPCClient {
	return &RunnableGRPCClient{context: context, client: client}
}

// Runnable interface
func (self *RunnableGRPCClient) Instantiate(config any) error {
	if config_, err := protobuf.NewStruct(config); err == nil {
		_, err := self.client.Instantiate(self.context, &api.Config{Config: config_})
		return err
	} else {
		return err
	}
}

// Runnable interface
func (self *RunnableGRPCClient) Interact(server util.Interactor, first *api.Interaction) error {
	if client, err := self.client.Interact(self.context); err == nil {
		return util.InteractRelay(server, client, first, log)
	} else {
		return err
	}
}

//
// RunnablePlugin
//

type RunnablePlugin struct {
	plugin.Plugin

	implementation Runnable // only for servers
}

// plugin.GRPCPlugin interface
func (self *RunnablePlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	api.RegisterPluginServer(server, NewRunnableGRPCServer(self.implementation))
	return nil
}

// plugin.GRPCPlugin interface
func (p *RunnablePlugin) GRPCClient(context contextpkg.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	return NewRunnableGRPCClient(context, api.NewPluginClient(client)), nil
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
			"runnable": new(RunnablePlugin),
		},
		Cmd:              exec.Command(command),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
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
		if runnable, err := protocol.Dispense("runnable"); err == nil {
			return runnable.(Runnable), nil
		} else {
			protocol.Close()
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
	plugin plugin.Plugin
}

func NewRunnableServer(implementation Runnable) *RunnableServer {
	return &RunnableServer{
		plugin: &RunnablePlugin{implementation: implementation},
	}
}

func (self *RunnableServer) Start() {
	var config = plugin.ServeConfig{
		Plugins: map[string]plugin.Plugin{
			"runnable": self.plugin,
		},
		GRPCServer:      plugin.DefaultGRPCServer,
		HandshakeConfig: handshakeConfig,
		Logger:          sink.NewHCLogger("khutulun.plugin.server", nil),
	}

	plugin.Serve(&config)
}
