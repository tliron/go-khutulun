package conductor

import (
	contextpkg "context"
	"fmt"
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
	api.UnimplementedConductorServer

	Protocol string
	Address  string
	Port     int

	grpcServer *grpc.Server
	conductor  *Conductor
}

func NewGRPC(conductor *Conductor, protocol string, address string, port int) *GRPC {
	return &GRPC{
		Protocol:  protocol,
		Address:   address,
		Port:      port,
		conductor: conductor,
	}
}

func (self *GRPC) Start() error {
	self.grpcServer = grpc.NewServer()
	api.RegisterConductorServer(self.grpcServer, self)

	if listener, err := newListener(self.Protocol, self.Address, self.Port); err == nil {
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

func (self *GRPC) Stop() {
	if self.grpcServer != nil {
		self.grpcServer.Stop()
	}
}

// api.ConductorServer interface
func (self *GRPC) GetVersion(context contextpkg.Context, empty *emptypb.Empty) (*api.Version, error) {
	grpcLog.Info("getVersion()")

	return &version, nil
}

// api.ConductorServer interface
func (self *GRPC) ListHosts(empty *emptypb.Empty, server api.Conductor_ListHostsServer) error {
	grpcLog.Info("listHosts()")

	if self.conductor.cluster != nil {
		for _, member := range self.conductor.cluster.ListHosts() {
			server.Send(&api.HostIdentifier{
				Name:    member.name,
				Address: member.address,
			})
		}
		return nil
	} else {
		return statuspkg.Error(codes.Aborted, "cluster not enabled")
	}
}

// api.ConductorServer interface
func (self *GRPC) AddHost(context contextpkg.Context, identifier *api.HostIdentifier) (*emptypb.Empty, error) {
	grpcLog.Infof("addHost(%q)", identifier.Address)

	if self.conductor.cluster != nil {
		if err := self.conductor.cluster.AddHosts([]string{identifier.Address}); err == nil {
			return new(emptypb.Empty), nil
		} else {
			return new(emptypb.Empty), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
		}
	} else {
		return new(emptypb.Empty), statuspkg.Error(codes.Aborted, "cluster not enabled")
	}
}

// api.ConductorServer interface
func (self *GRPC) ListNamespaces(empty *emptypb.Empty, server api.Conductor_ListNamespacesServer) error {
	grpcLog.Info("listNamespaces()")

	if namespaces, err := self.conductor.ListNamespaces(); err == nil {
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

// api.ConductorServer interface
func (self *GRPC) ListPackages(listPackages *api.ListPackages, server api.Conductor_ListPackagesServer) error {
	grpcLog.Infof("listPackages(%q, %q)", listPackages.Namespace, listPackages.Type.Name)

	if identifiers, err := self.conductor.ListPackages(listPackages.Namespace, listPackages.Type.Name); err == nil {
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

// api.ConductorServer interface
func (self *GRPC) ListPackageFiles(identifier *api.PackageIdentifier, server api.Conductor_ListPackageFilesServer) error {
	grpcLog.Infof("listPackageFiles(%q, %q, %q)", identifier.Namespace, identifier.Type.Name, identifier.Name)

	if packageFiles, err := self.conductor.ListPackageFiles(identifier.Namespace, identifier.Type.Name, identifier.Name); err == nil {
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

// api.ConductorServer interface
func (self *GRPC) GetPackageFiles(getPackageFiles *api.GetPackageFiles, server api.Conductor_GetPackageFilesServer) error {
	grpcLog.Infof("getPackageFiles(%q, %q, %q)", getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name)

	if lock, err := self.conductor.lockPackage(getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name, false); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", grpcLog)

		buffer := make([]byte, BUFFER_SIZE)
		dir := self.conductor.getPackageDir(getPackageFiles.Identifier.Namespace, getPackageFiles.Identifier.Type.Name, getPackageFiles.Identifier.Name)

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

// api.ConductorServer interface
func (self *GRPC) SetPackageFiles(server api.Conductor_SetPackageFilesServer) error {
	grpcLog.Info("setPackageFiles()")

	if first, err := server.Recv(); err == nil {
		if first.Start != nil {
			namespace := first.Start.Identifier.Namespace
			type_ := first.Start.Identifier.Type.Name
			name := first.Start.Identifier.Name
			if lock, err := self.conductor.lockPackage(namespace, type_, name, true); err == nil {
				defer logging.CallAndLogError(lock.Unlock, "unlock", grpcLog)

				var file *os.File
				for {
					if content, err := server.Recv(); err == nil {
						if content.Start != nil {
							if file != nil {
								if err := file.Close(); err != nil {
									grpcLog.Errorf("file close: %s", err.Error())
								}
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
							path := filepath.Join(self.conductor.getPackageDir(namespace, type_, name), content.File.Path)
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
							if err := file.Close(); err != nil {
								grpcLog.Errorf("file close: %s", err.Error())
							}
							return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
						}
					} else {
						if err == io.EOF {
							break
						} else {
							if file != nil {
								if err := file.Close(); err != nil {
									grpcLog.Errorf("file close: %s", err.Error())
								}
							}
							return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
						}
					}
				}

				if file != nil {
					if err := file.Close(); err != nil {
						grpcLog.Errorf("file close: %s", err.Error())
					}
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

// api.ConductorServer interface
func (self *GRPC) RemovePackage(context contextpkg.Context, packageIdentifer *api.PackageIdentifier) (*emptypb.Empty, error) {
	grpcLog.Infof("removePackage(%q, %q, %q)", packageIdentifer.Namespace, packageIdentifer.Type.Name, packageIdentifer.Name)

	if err := self.conductor.DeletePackage(packageIdentifer.Namespace, packageIdentifer.Type.Name, packageIdentifer.Name); err == nil {
		return new(emptypb.Empty), nil
	} else {
		return new(emptypb.Empty), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.ConductorServer interface
func (self *GRPC) DeployService(context contextpkg.Context, deployService *api.DeployService) (*emptypb.Empty, error) {
	grpcLog.Infof("deployService(%q, %q)", deployService.Service.Name, deployService.Template.Name)

	if err := self.conductor.DeployService(deployService.Template.Namespace, deployService.Template.Name, deployService.Service.Namespace, deployService.Service.Name); err == nil {
		return new(emptypb.Empty), nil
	} else {
		return new(emptypb.Empty), statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

// api.ConductorServer interface
func (self *GRPC) ListResources(listResources *api.ListResources, server api.Conductor_ListResourcesServer) error {
	grpcLog.Info("listResources()")

	if identifiers, err := self.conductor.ListResources(listResources.Service.Namespace, listResources.Service.Name, listResources.Type); err == nil {
		for _, identifier := range identifiers {
			identifier_ := api.ResourceIdentifier{
				Service: &api.ServiceIdentifier{
					Namespace: identifier.Namespace,
					Name:      identifier.Service,
				},
				Type: identifier.Type,
				Name: identifier.Name,
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

// api.ConductorServer interface
func (self *GRPC) Interact(server api.Conductor_InteractServer) error {
	grpcLog.Info("interact()")

	return util.Interact(server, map[string]util.InteractFunc{
		"host": func(start *api.Interaction_Start) error {
			if len(start.Identifier) != 2 {
				return statuspkg.Errorf(codes.InvalidArgument, "malformed identifier for host: %s", start.Identifier)
			}

			host := start.Identifier[1]

			command := util.NewCommand(start, grpcLog)

			var relay string
			if self.conductor.cluster != nil {
				if self.conductor.host != host {
					for _, node := range self.conductor.cluster.cluster.Members() {
						if node.Name == host {
							// TODO: we need the gRPC info for the host
							relay = fmt.Sprintf("[%s]:%d", host, self.Port)
						}
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
			command := self.conductor.getPackageMainFile("common", "plugin", name)

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
