package delegate

import (
	contextpkg "context"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/util"
	"github.com/tliron/kutil/protobuf"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//
// DelegateGRPCServer
//

type DelegateGRPCServer struct {
	api.UnimplementedDelegateServer

	implementation Delegate
}

func NewDelegateGRPCServer(implementation Delegate) *DelegateGRPCServer {
	return &DelegateGRPCServer{implementation: implementation}
}

// api.DelegateServer interface
func (self *DelegateGRPCServer) Instantiate(context contextpkg.Context, config *api.Config) (*emptypb.Empty, error) {
	if err := self.implementation.Instantiate(config.Config.AsMap()); err == nil {
		return new(emptypb.Empty), nil
	} else {
		return new(emptypb.Empty), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.DelegateServer interface
func (self *DelegateGRPCServer) Interact(server api.Delegate_InteractServer) error {
	return util.Interact(server, map[string]util.InteractFunc{
		"runnable": func(start *api.Interaction_Start) error {
			return self.implementation.Interact(server, start)
		},
	})
}

//
// DelegateGRPCClient
//

type DelegateGRPCClient struct {
	context contextpkg.Context
	client  api.DelegateClient
}

func NewDelegateGRPCClient(context contextpkg.Context, client api.DelegateClient) *DelegateGRPCClient {
	return &DelegateGRPCClient{context: context, client: client}
}

// Delegate interface
func (self *DelegateGRPCClient) Instantiate(config any) error {
	if config_, err := protobuf.NewStruct(config); err == nil {
		_, err := self.client.Instantiate(self.context, &api.Config{Config: config_})
		return err
	} else {
		return err
	}
}

// Delegate interface
func (self *DelegateGRPCClient) Interact(server util.GRPCInteractor, start *api.Interaction_Start) error {
	if client, err := self.client.Interact(self.context); err == nil {
		return util.InteractRelay(server, client, start, log)
	} else {
		return err
	}
}
