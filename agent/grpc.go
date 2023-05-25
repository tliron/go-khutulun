package agent

import (
	contextpkg "context"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"

	"github.com/tliron/commonlog"
	"github.com/tliron/khutulun/api"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/khutulun/sdk"
	"github.com/tliron/kutil/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const BUFFER_SIZE = 65536

var version = api.Version{Version: "0.1.0"}

//
// GRPC
//

type GRPC struct {
	api.UnimplementedAgentServer

	Protocol string
	Address  string
	Port     int
	Local    bool

	grpcServer *grpc.Server
	agent      *Agent
}

func NewGRPC(agent *Agent, protocol string, address string, port int) *GRPC {
	return &GRPC{
		Protocol: protocol,
		Address:  address,
		Port:     port,
		Local:    true,
		agent:    agent,
	}
}

func (self *GRPC) Start() error {
	self.grpcServer = grpc.NewServer()
	api.RegisterAgentServer(self.grpcServer, self)

	var err error
	if self.Address, err = util.ToReachableIPAddress(self.Address); err != nil {
		return err
	}

	start := func(address string) error {
		if listener, err := net.Listen(self.Protocol, util.JoinIPAddressPort(address, self.Port)); err == nil {
			grpcLog.Noticef("starting server on %s", listener.Addr().String())
			go func() {
				if err := self.grpcServer.Serve(listener); err != nil {
					grpcLog.Errorf("%s", err.Error())
				}
			}()
			return nil
		} else {
			return err
		}
	}

	if self.Local {
		if (self.Protocol == "tcp") || (self.Protocol == "tcp6") {
			if err := start("::1"); err != nil {
				return err
			}
		} else if self.Protocol == "tcp4" {
			if err := start("127.0.0.1"); err != nil {
				return err
			}
		}
	}

	return start(self.Address)
}

func (self *GRPC) Stop() {
	if self.grpcServer != nil {
		self.grpcServer.Stop()
	}
}

// api.AgentServer interface
func (self *GRPC) GetVersion(context contextpkg.Context, empty *emptypb.Empty) (*api.Version, error) {
	grpcLog.Info("getVersion()")

	return &version, nil
}

// api.AgentServer interface
func (self *GRPC) ListHosts(empty *emptypb.Empty, server api.Agent_ListHostsServer) error {
	grpcLog.Info("listHosts()")

	if self.agent.gossip != nil {
		for _, host := range self.agent.gossip.ListHosts() {
			server.Send(&api.HostIdentifier{
				Name:        host.Name,
				GrpcAddress: host.GRPCAddress,
			})
		}
		return nil
	} else {
		return statuspkg.Error(codes.Aborted, "gossip not enabled")
	}
}

// api.AgentServer interface
func (self *GRPC) AddHost(context contextpkg.Context, addHost *api.AddHost) (*emptypb.Empty, error) {
	grpcLog.Infof("addHost(%q)", addHost.GossipAddress)

	if self.agent.gossip != nil {
		if err := self.agent.gossip.AddHosts([]string{addHost.GossipAddress}); err == nil {
			return new(emptypb.Empty), nil
		} else {
			return new(emptypb.Empty), sdk.GRPCAborted(err)
		}
	} else {
		return new(emptypb.Empty), statuspkg.Error(codes.Aborted, "gossip not enabled")
	}
}

// api.AgentServer interface
func (self *GRPC) ListNamespaces(empty *emptypb.Empty, server api.Agent_ListNamespacesServer) error {
	grpcLog.Info("listNamespaces()")

	if namespaces, err := self.agent.state.ListNamespaces(); err == nil {
		for _, namespace := range namespaces {
			if err := server.Send(&api.Namespace{Name: namespace}); err != nil {
				return sdk.GRPCAborted(err)
			}
		}
		return nil
	} else {
		return sdk.GRPCAborted(err)
	}
}

// api.AgentServer interface
func (self *GRPC) ListPackages(listPackages *api.ListPackages, server api.Agent_ListPackagesServer) error {
	grpcLog.Infof("listPackages(%q, %q)", listPackages.Namespace, listPackages.Type.Name)

	if identifiers, err := self.agent.state.ListPackages(listPackages.Namespace, listPackages.Type.Name); err == nil {
		for _, identifier := range identifiers {
			identifier_ := api.PackageIdentifier{
				Namespace: identifier.Namespace,
				Type:      &api.PackageType{Name: identifier.Type},
				Name:      identifier.Name,
			}

			if err := server.Send(&identifier_); err != nil {
				return sdk.GRPCAborted(err)
			}
		}
	} else {
		return sdk.GRPCAborted(err)
	}

	return nil
}

// api.AgentServer interface
func (self *GRPC) ListPackageFiles(identifier *api.PackageIdentifier, server api.Agent_ListPackageFilesServer) error {
	grpcLog.Infof("listPackageFiles(%q, %q, %q)", identifier.Namespace, identifier.Type.Name, identifier.Name)

	if packageFiles, err := self.agent.state.ListPackageFiles(identifier.Namespace, identifier.Type.Name, identifier.Name); err == nil {
		for _, packageFile := range packageFiles {
			if err := server.Send(&api.PackageFile{
				Path:       packageFile.Path,
				Executable: packageFile.Executable,
			}); err != nil {
				return sdk.GRPCAborted(err)
			}
		}
	} else {
		return sdk.GRPCAborted(err)
	}

	return nil
}

// api.AgentServer interface
func (self *GRPC) GetPackageFiles(getPackageFiles *api.GetPackageFiles, server api.Agent_GetPackageFilesServer) error {
	grpcLog.Infof("getPackageFiles(%q, %q, %q)", getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name)

	if lock, err := self.agent.state.LockPackage(getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name, false); err == nil {
		defer commonlog.CallAndLogError(lock.Unlock, "unlock", grpcLog)

		buffer := make([]byte, BUFFER_SIZE)
		dir := self.agent.state.GetPackageDir(getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name)

		for _, path := range getPackageFiles.Paths {
			if reader, err := self.agent.OpenFile(filepath.Join(dir, path), getPackageFiles.Coerce); err == nil {
				reader = util.NewContextualReadCloser(server.Context(), reader)
				for {
					count, err := reader.Read(buffer)
					if count > 0 {
						content := api.PackageContent{Bytes: buffer[:count]}
						if err := server.Send(&content); err != nil {
							if err := reader.Close(); err != nil {
								grpcLog.Errorf("file close: %s", err.Error())
							}
							return sdk.GRPCAborted(err)
						}
					}
					if err != nil {
						if err == io.EOF {
							break
						} else {
							if err := reader.Close(); err != nil {
								grpcLog.Errorf("file close: %s", err.Error())
							}
							return sdk.GRPCAborted(err)
						}
					}
				}

				commonlog.CallAndLogError(reader.Close, "file close", grpcLog)
			} else {
				return sdk.GRPCAborted(err)
			}
		}

		return nil
	} else {
		return sdk.GRPCAborted(err)
	}
}

// api.AgentServer interface
func (self *GRPC) SetPackageFiles(server api.Agent_SetPackageFilesServer) error {
	grpcLog.Info("setPackageFiles()")

	if first, err := server.Recv(); err == nil {
		if first.Start != nil {
			namespace := first.Start.Identifier.Namespace
			type_ := first.Start.Identifier.Type.Name
			name := first.Start.Identifier.Name
			if lock, err := self.agent.state.LockPackage(namespace, type_, name, true); err == nil {
				defer commonlog.CallAndLogError(lock.Unlock, "unlock", grpcLog)

				var file *os.File
				for {
					if content, err := server.Recv(); err == nil {
						if content.Start != nil {
							if file != nil {
								commonlog.CallAndLogError(file.Close, "file close", grpcLog)
							}
							return statuspkg.Error(codes.InvalidArgument, "received more than one message with \"start\"")
						}

						if content.File != nil {
							if content.File.Path == sdk.LOCK_FILE {
								// TODO
							}

							if file != nil {
								if err := file.Close(); err != nil {
									return sdk.GRPCAborted(err)
								}
								file = nil
							}
							path := filepath.Join(self.agent.state.GetPackageDir(namespace, type_, name), content.File.Path)
							if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
								return sdk.GRPCAborted(err)
							}

							var mode fs.FileMode = 0666
							if content.File.Executable {
								mode = 0777
							}

							if file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode); err != nil {
								return sdk.GRPCAborted(err)
							}
						}

						if file == nil {
							return statuspkg.Errorf(codes.Aborted, "message must container \"file\"")
						}

						if _, err := file.Write(content.Bytes); err != nil {
							commonlog.CallAndLogError(file.Close, "file close", grpcLog)
							return sdk.GRPCAborted(err)
						}
					} else {
						if err == io.EOF {
							break
						} else {
							if file != nil {
								commonlog.CallAndLogError(file.Close, "file close", grpcLog)
							}
							return sdk.GRPCAborted(err)
						}
					}
				}

				if file != nil {
					commonlog.CallAndLogError(file.Close, "file close", grpcLog)
				}

				return nil
			} else {
				return sdk.GRPCAborted(err)
			}
		} else {
			return statuspkg.Error(codes.InvalidArgument, "first message must contain \"start\"")
		}
	} else {
		return sdk.GRPCAborted(err)
	}
}

// api.AgentServer interface
func (self *GRPC) RemovePackage(context contextpkg.Context, packageIdentifer *api.PackageIdentifier) (*emptypb.Empty, error) {
	grpcLog.Infof("removePackage(%q, %q, %q)", packageIdentifer.Namespace, packageIdentifer.Type.Name, packageIdentifer.Name)

	if err := self.agent.state.DeletePackage(packageIdentifer.Namespace, packageIdentifer.Type.Name, packageIdentifer.Name); err == nil {
		return new(emptypb.Empty), nil
	} else {
		return new(emptypb.Empty), sdk.GRPCAborted(err)
	}
}

// api.AgentServer interface
func (self *GRPC) DeployService(context contextpkg.Context, deployService *api.DeployService) (*emptypb.Empty, error) {
	grpcLog.Infof("deployService(%q, %q, %q, %q, %t)", deployService.Template.Namespace, deployService.Template.Name, deployService.Service.Name, deployService.Template.Name, deployService.Async)

	if err := self.agent.DeployService(context, deployService.Template.Namespace, deployService.Template.Name, deployService.Service.Namespace, deployService.Service.Name, deployService.Async); err == nil {
		return new(emptypb.Empty), nil
	} else {
		return new(emptypb.Empty), sdk.GRPCAborted(err)
	}
}

// api.AgentServer interface
func (self *GRPC) ListResources(listResources *api.ListResources, server api.Agent_ListResourcesServer) error {
	grpcLog.Infof("listResources(%q, %q, %q)", listResources.Service.Namespace, listResources.Service.Name, listResources.Type)

	if identifiers, err := self.agent.ListResources(server.Context(), listResources.Service.Namespace, listResources.Service.Name, listResources.Type); err == nil {
		for _, identifier := range identifiers {
			identifier_ := api.ResourceIdentifier{
				Service: &api.ServiceIdentifier{
					Namespace: identifier.Namespace,
					Name:      identifier.Service,
				},
				Type: identifier.Type,
				Name: identifier.Name,
				Host: identifier.Host,
			}

			if err := server.Send(&identifier_); err != nil {
				return sdk.GRPCAborted(err)
			}
		}
	} else {
		return sdk.GRPCAborted(err)
	}

	return nil
}

// api.AgentServer interface
func (self *GRPC) Interact(server api.Agent_InteractServer) error {
	grpcLog.Info("interact()")

	relay := func(host string, start *api.Interaction_Start) error {
		if host_ := self.agent.gossip.GetHost(host); host_ != nil {
			client, err := clientpkg.NewClient(host_.GRPCAddress)
			if err != nil {
				return err
			}
			defer client.Close()

			grpcLog.Infof("relay interaction to %s", host_)
			err = client.InteractRelay(server, start)
			grpcLog.Info("interaction ended")
			return err
		} else {
			return statuspkg.Errorf(codes.Aborted, "host not found: %s", host)
		}
	}

	return sdk.Interact(server, map[string]sdk.InteractFunc{
		"host": func(start *api.Interaction_Start) error {
			if len(start.Identifier) != 2 {
				return statuspkg.Errorf(codes.InvalidArgument, "malformed identifier for host: %s", start.Identifier)
			}

			host := start.Identifier[1]

			command := sdk.NewCommand(start, grpcLog)

			if self.agent.gossip != nil {
				if self.agent.host != host {
					return relay(host, start)
				}
			}

			return sdk.StartCommand(command, server, grpcLog)
		},

		"activity": func(start *api.Interaction_Start) error {
			// TODO: find host for activity and relay if necessary

			namespace := start.Identifier[1]
			serviceName := start.Identifier[2]
			resourceName := start.Identifier[3]

			if lock, clout, err := self.agent.state.OpenServiceClout(server.Context(), namespace, serviceName, self.agent.urlContext); err == nil {
				commonlog.CallAndLogError(lock.Unlock, "unlock", delegateLog)
				if clout, err = self.agent.CoerceClout(clout, false); err == nil {
					delegates := self.agent.NewDelegates()
					defer delegates.Release()
					delegates.Fill(namespace, clout)

					for _, delegate := range delegates.All() {
						if resources, err := delegate.ListResources(namespace, serviceName, clout); err == nil {
							var found bool
							for _, resource := range resources {
								if resource.Name == resourceName {
									if resource.Host != self.agent.host {
										return relay(resource.Host, start)
									}
									found = true
									break
								}
							}

							if found {
								return delegate.Interact(server, start)
							}
						} else {
							return sdk.GRPCAborted(err)
						}
					}
				} else {
					return sdk.GRPCAborted(err)
				}
			} else {
				return sdk.GRPCAborted(err)
			}

			return sdk.GRPCAbortedf("activity not found: %s/%s->%s", namespace, serviceName, resourceName)
		},
	})
}
