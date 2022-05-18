package delegate

import (
	contextpkg "context"
	"strings"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/format"
	cloutpkg "github.com/tliron/puccini/clout"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
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
func (self *DelegateGRPCServer) ProcessService(context contextpkg.Context, processService *api.ProcessService) (*api.ProcessServiceResult, error) {
	if clout, err := cloutpkg.Read(strings.NewReader(processService.Clout.Yaml), "yaml"); err == nil {
		if coercedClout, err := cloutpkg.Read(strings.NewReader(processService.Coerced.Yaml), "yaml"); err == nil {
			if clout_, err := self.implementation.ProcessService(processService.Service.Namespace, processService.Service.Name, processService.Phase, clout, coercedClout); err == nil {
				if clout_ != nil {
					if clout__, err := format.Encode(clout_, "yaml", " ", false); err == nil {
						return &api.ProcessServiceResult{
							Clout: &api.Clout{Yaml: clout__},
						}, nil
					} else {
						return new(api.ProcessServiceResult), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
					}
				} else {
					return new(api.ProcessServiceResult), nil
				}
			} else {
				return nil, statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		} else {
			return nil, statuspkg.Errorf(codes.Aborted, "%s", err.Error())
		}
	} else {
		return nil, statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.DelegateServer interface
func (self *DelegateGRPCServer) Interact(server api.Delegate_InteractServer) error {
	return sdk.Interact(server, map[string]sdk.InteractFunc{
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
func (self *DelegateGRPCClient) ProcessService(namespace string, serviceName string, phase string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, error) {
	if clout_, err := format.Encode(clout, "yaml", " ", false); err == nil {
		if coercedClout_, err := format.Encode(coercedClout, "yaml", " ", false); err == nil {
			processService := api.ProcessService{
				Service: &api.ServiceIdentifier{
					Namespace: namespace,
					Name:      serviceName,
				},
				Phase:   phase,
				Clout:   &api.Clout{Yaml: clout_},
				Coerced: &api.Clout{Yaml: coercedClout_},
			}
			if result, err := self.client.ProcessService(self.context, &processService); err == nil {
				if (result != nil) && (result.Clout != nil) {
					if clout__, err := cloutpkg.Read(strings.NewReader(result.Clout.Yaml), "yaml"); err == nil {
						return clout__, nil
					} else {
						return nil, err
					}
				} else {
					return nil, nil
				}
			} else {
				return nil, sdk.UnpackGRPCError(err)
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// Delegate interface
func (self *DelegateGRPCClient) Interact(server sdk.GRPCInteractor, start *api.Interaction_Start) error {
	if client, err := self.client.Interact(self.context); err == nil {
		return sdk.InteractRelay(server, client, start, log)
	} else {
		return err
	}
}
