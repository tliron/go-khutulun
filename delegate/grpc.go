package delegate

import (
	contextpkg "context"
	"io"
	"net"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/util"
	cloutpkg "github.com/tliron/puccini/clout"
	"google.golang.org/grpc"
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

func (self *DelegateGRPCServer) Start(protocol string, address string, port int) error {
	grpcServer := grpc.NewServer()
	api.RegisterDelegateServer(grpcServer, self)

	if listener, err := net.Listen(protocol, util.JoinIPAddressPort(address, port)); err == nil {
		log.Noticef("starting server on %s", listener.Addr().String())
		return grpcServer.Serve(listener)
	} else {
		return err
	}
}

// api.DelegateServer interface
func (self *DelegateGRPCServer) ListResources(listResources *api.DelegateListResources, server api.Delegate_ListResourcesServer) error {
	if coercedClout, err := CloutFromAPI(listResources.CoercedClout); err == nil {
		if resources, err := self.implementation.ListResources(listResources.Service.Namespace, listResources.Service.Name, coercedClout); err == nil {
			for _, resource := range resources {
				resource_ := api.ResourceIdentifier{
					Service: &api.ServiceIdentifier{
						Namespace: listResources.Service.Namespace,
						Name:      listResources.Service.Name,
					},
					Type: resource.Type,
					Name: resource.Name,
					Host: resource.Host,
				}
				if err := server.Send(&resource_); err != nil {
					return sdk.GRPCAborted(err)
				}
			}
		} else {
			return sdk.GRPCAborted(err)
		}
	} else {
		return sdk.GRPCAborted(err)
	}

	return nil
}

// api.DelegateServer interface
func (self *DelegateGRPCServer) ProcessService(context contextpkg.Context, processService *api.ProcessService) (*api.ProcessServiceResult, error) {
	if clout, err := CloutFromAPI(processService.Clout); err == nil {
		if coercedClout, err := CloutFromAPI(processService.CoercedClout); err == nil {
			if clout_, next, err := self.implementation.ProcessService(processService.Service.Namespace, processService.Service.Name, processService.Phase, clout, coercedClout); err == nil {
				var result api.ProcessServiceResult
				if result.Clout, err = CloutToAPI(clout_, false); err != nil {
					return new(api.ProcessServiceResult), sdk.GRPCAborted(err)
				}
				result.Next = NextsToAPI(next)
				return &result, nil
			} else {
				return new(api.ProcessServiceResult), sdk.GRPCAborted(err)
			}
		} else {
			return new(api.ProcessServiceResult), sdk.GRPCAborted(err)
		}
	} else {
		return new(api.ProcessServiceResult), sdk.GRPCAborted(err)
	}
}

// api.DelegateServer interface
func (self *DelegateGRPCServer) Interact(server api.Delegate_InteractServer) error {
	return sdk.Interact(server, map[string]sdk.InteractFunc{
		"activity": func(start *api.Interaction_Start) error {
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
func (self *DelegateGRPCClient) ListResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout) ([]Resource, error) {
	if coercedClout_, err := CloutToAPI(coercedClout, false); err == nil {
		listResources := api.DelegateListResources{
			Service: &api.ServiceIdentifier{
				Namespace: namespace,
				Name:      serviceName,
			},
			CoercedClout: coercedClout_,
		}
		if client, err := self.client.ListResources(self.context, &listResources); err == nil {
			var resources []Resource
			for {
				if resource, err := client.Recv(); err == nil {
					resources = append(resources, Resource{
						Type: resource.Type,
						Name: resource.Name,
						Host: resource.Host,
					})
				} else if err == io.EOF {
					break
				} else {
					return nil, sdk.UnpackGRPCError(err)
				}
			}
			return resources, nil
		} else {
			return nil, sdk.UnpackGRPCError(err)
		}
	} else {
		return nil, err
	}
}

// Delegate interface
func (self *DelegateGRPCClient) ProcessService(namespace string, serviceName string, phase string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []Next, error) {
	if clout_, err := CloutToAPI(clout, false); err == nil {
		if coercedClout_, err := CloutToAPI(coercedClout, false); err == nil {
			processService := api.ProcessService{
				Service: &api.ServiceIdentifier{
					Namespace: namespace,
					Name:      serviceName,
				},
				Phase:        phase,
				Clout:        clout_,
				CoercedClout: coercedClout_,
			}
			if result, err := self.client.ProcessService(self.context, &processService); err == nil {
				if result != nil {
					var clout__ *cloutpkg.Clout
					if clout__, err = CloutFromAPI(result.Clout); err != nil {
						return nil, nil, err
					}
					return clout__, NextsFromAPI(result.Next), nil
				} else {
					return nil, nil, nil
				}
			} else {
				return nil, nil, sdk.UnpackGRPCError(err)
			}
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, err
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
