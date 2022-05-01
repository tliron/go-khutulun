package host

import (
	contextpkg "context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tliron/khutulun/api"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/khutulun/util"
	"github.com/tliron/kutil/logging"
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
	api.UnimplementedHostServer

	Protocol string
	Address  string
	Port     int
	Local    bool

	grpcServer *grpc.Server
	host       *Host
}

func NewGRPC(host *Host, protocol string, address string, port int) *GRPC {
	return &GRPC{
		Protocol: protocol,
		Address:  address,
		Port:     port,
		Local:    true,
		host:     host,
	}
}

func (self *GRPC) Start() error {
	self.grpcServer = grpc.NewServer()
	api.RegisterHostServer(self.grpcServer, self)

	var err error
	if self.Address, err = toReachableAddress(self.Address); err != nil {
		return err
	}

	start := func(address string) error {
		if listener, err := newListener(self.Protocol, address, self.Port); err == nil {
			grpcLog.Noticef("starting server on: %s", listener.Addr().String())
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
		if err := start("localhost"); err != nil {
			return err
		}
	}

	if err := start(self.Address); err != nil {
		return err
	}

	return nil
}

func (self *GRPC) Stop() {
	if self.grpcServer != nil {
		self.grpcServer.Stop()
	}
}

// api.HostServer interface
func (self *GRPC) GetVersion(context contextpkg.Context, empty *emptypb.Empty) (*api.Version, error) {
	grpcLog.Info("getVersion()")

	return &version, nil
}

// api.HostServer interface
func (self *GRPC) ListHosts(empty *emptypb.Empty, server api.Host_ListHostsServer) error {
	grpcLog.Info("listHosts()")

	if self.host.gossip != nil {
		for _, host := range self.host.gossip.ListHosts() {
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

// api.HostServer interface
func (self *GRPC) AddHost(context contextpkg.Context, addHost *api.AddHost) (*emptypb.Empty, error) {
	grpcLog.Infof("addHost(%q)", addHost.GossipAddress)

	if self.host.gossip != nil {
		if err := self.host.gossip.AddHosts([]string{addHost.GossipAddress}); err == nil {
			return new(emptypb.Empty), nil
		} else {
			return new(emptypb.Empty), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
		}
	} else {
		return new(emptypb.Empty), statuspkg.Error(codes.Aborted, "gossip not enabled")
	}
}

// api.HostServer interface
func (self *GRPC) ListNamespaces(empty *emptypb.Empty, server api.Host_ListNamespacesServer) error {
	grpcLog.Info("listNamespaces()")

	if namespaces, err := self.host.ListNamespaces(); err == nil {
		for _, namespace := range namespaces {
			if err := server.Send(&api.Namespace{Name: namespace}); err != nil {
				return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		}
		return nil
	} else {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.HostServer interface
func (self *GRPC) ListPackages(listPackages *api.ListPackages, server api.Host_ListPackagesServer) error {
	grpcLog.Infof("listPackages(%q, %q)", listPackages.Namespace, listPackages.Type.Name)

	if identifiers, err := self.host.ListPackages(listPackages.Namespace, listPackages.Type.Name); err == nil {
		for _, identifier := range identifiers {
			identifier_ := api.PackageIdentifier{
				Namespace: identifier.Namespace,
				Type:      &api.PackageType{Name: identifier.Type},
				Name:      identifier.Name,
			}

			if err := server.Send(&identifier_); err != nil {
				return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		}
	} else {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}

	return nil
}

// api.HostServer interface
func (self *GRPC) ListPackageFiles(identifier *api.PackageIdentifier, server api.Host_ListPackageFilesServer) error {
	grpcLog.Infof("listPackageFiles(%q, %q, %q)", identifier.Namespace, identifier.Type.Name, identifier.Name)

	if packageFiles, err := self.host.ListPackageFiles(identifier.Namespace, identifier.Type.Name, identifier.Name); err == nil {
		for _, packageFile := range packageFiles {
			if err := server.Send(&api.PackageFile{
				Path:       packageFile.Path,
				Executable: packageFile.Executable,
			}); err != nil {
				return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		}
	} else {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}

	return nil
}

// api.HostServer interface
func (self *GRPC) GetPackageFiles(getPackageFiles *api.GetPackageFiles, server api.Host_GetPackageFilesServer) error {
	grpcLog.Infof("getPackageFiles(%q, %q, %q)", getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name)

	if lock, err := self.host.lockPackage(getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name, false); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", grpcLog)

		buffer := make([]byte, BUFFER_SIZE)
		dir := self.host.getPackageDir(getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name)

		for _, path := range getPackageFiles.Paths {
			if file, err := os.Open(filepath.Join(dir, path)); err == nil {
				for {
					count, err := file.Read(buffer)
					if count > 0 {
						content := api.PackageContent{Bytes: buffer[:count]}
						if err := server.Send(&content); err != nil {
							if err := file.Close(); err != nil {
								grpcLog.Errorf("file close: %s", err.Error())
							}
							return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
						}
					}
					if err != nil {
						if err == io.EOF {
							break
						} else {
							if err := file.Close(); err != nil {
								grpcLog.Errorf("file close: %s", err.Error())
							}
							return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
						}
					}
				}

				if err := file.Close(); err != nil {
					grpcLog.Errorf("file close: %s", err.Error())
				}
			} else {
				return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		}

		return nil
	} else {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.HostServer interface
func (self *GRPC) SetPackageFiles(server api.Host_SetPackageFilesServer) error {
	grpcLog.Info("setPackageFiles()")

	if first, err := server.Recv(); err == nil {
		if first.Start != nil {
			namespace := first.Start.Identifier.Namespace
			type_ := first.Start.Identifier.Type.Name
			name := first.Start.Identifier.Name
			if lock, err := self.host.lockPackage(namespace, type_, name, true); err == nil {
				defer logging.CallAndLogError(lock.Unlock, "unlock", grpcLog)

				var file *os.File
				for {
					if content, err := server.Recv(); err == nil {
						if content.Start != nil {
							if file != nil {
								logging.CallAndLogError(file.Close, "file close", grpcLog)
							}
							return statuspkg.Error(codes.InvalidArgument, "received more than one message with \"start\"")
						}

						if content.File != nil {
							if content.File.Path == LOCK_FILE {
								// TODO
							}

							if file != nil {
								if err := file.Close(); err != nil {
									return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
								}
								file = nil
							}
							path := filepath.Join(self.host.getPackageDir(namespace, type_, name), content.File.Path)
							if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
								return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
							}

							var mode fs.FileMode = 0666
							if content.File.Executable {
								mode = 0777
							}

							if file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode); err != nil {
								return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
							}
						}

						if file == nil {
							return statuspkg.Errorf(codes.Aborted, "message must container \"file\"")
						}

						if _, err := file.Write(content.Bytes); err != nil {
							logging.CallAndLogError(file.Close, "file close", grpcLog)
							return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
						}
					} else {
						if err == io.EOF {
							break
						} else {
							if file != nil {
								logging.CallAndLogError(file.Close, "file close", grpcLog)
							}
							return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
						}
					}
				}

				if file != nil {
					logging.CallAndLogError(file.Close, "file close", grpcLog)
				}

				return nil
			} else {
				return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		} else {
			return statuspkg.Error(codes.InvalidArgument, "first message must contain \"start\"")
		}
	} else {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.HostServer interface
func (self *GRPC) RemovePackage(context contextpkg.Context, packageIdentifer *api.PackageIdentifier) (*emptypb.Empty, error) {
	grpcLog.Infof("removePackage(%q, %q, %q)", packageIdentifer.Namespace, packageIdentifer.Type.Name, packageIdentifer.Name)

	if err := self.host.DeletePackage(packageIdentifer.Namespace, packageIdentifer.Type.Name, packageIdentifer.Name); err == nil {
		return new(emptypb.Empty), nil
	} else {
		return new(emptypb.Empty), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.HostServer interface
func (self *GRPC) DeployService(context contextpkg.Context, deployService *api.DeployService) (*emptypb.Empty, error) {
	grpcLog.Infof("deployService(%q, %q, %q, %q)", deployService.Template.Namespace, deployService.Template.Name, deployService.Service.Name, deployService.Template.Name)

	if err := self.host.DeployService(deployService.Template.Namespace, deployService.Template.Name, deployService.Service.Namespace, deployService.Service.Name); err == nil {
		return new(emptypb.Empty), nil
	} else {
		return new(emptypb.Empty), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.HostServer interface
func (self *GRPC) ListResources(listResources *api.ListResources, server api.Host_ListResourcesServer) error {
	grpcLog.Infof("listResources(%q, %q, %q)", listResources.Service.Namespace, listResources.Service.Name, listResources.Type)

	if identifiers, err := self.host.ListResources(listResources.Service.Namespace, listResources.Service.Name, listResources.Type); err == nil {
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
				return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		}
	} else {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}

	return nil
}

// api.HostServer interface
func (self *GRPC) Interact(server api.Host_InteractServer) error {
	grpcLog.Info("interact()")

	return util.Interact(server, map[string]util.InteractFunc{
		"host": func(start *api.Interaction_Start) error {
			if len(start.Identifier) != 2 {
				return statuspkg.Errorf(codes.InvalidArgument, "malformed identifier for host: %s", start.Identifier)
			}

			host := start.Identifier[1]

			command := util.NewCommand(start, grpcLog)

			var relay string
			if self.host.gossip != nil {
				if self.host.host != host {
					if host_ := self.host.gossip.GetHost(host); host_ != nil {
						relay = host_.GRPCAddress
					} else {
						return statuspkg.Errorf(codes.Aborted, "host not found: %s", host)
					}
				}
			}

			if relay == "" {
				return util.StartCommand(command, server, grpcLog)
			} else {
				client, err := clientpkg.NewClient(relay)
				if err != nil {
					return err
				}
				defer client.Close()

				grpcLog.Infof("relay interaction to %s", relay)
				err = client.InteractRelay(server, start)
				grpcLog.Info("interaction ended")
				if err == nil {
					return nil
				} else {
					return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
				}
			}
		},

		"runnable": func(start *api.Interaction_Start) error {
			// TODO: find host for runnable and relay if necessary

			name := "runnable.podman"
			command := self.host.getPackageMainFile("common", "plugin", name)

			client := plugin.NewRunnableClient(name, command)
			defer client.Close()

			if runnable, err := client.Runnable(); err == nil {
				if err := runnable.Interact(server, start); err == nil {
					return nil
				} else {
					return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
				}
			} else {
				return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		},
	})
}
