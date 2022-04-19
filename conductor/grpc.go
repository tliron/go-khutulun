package conductor

import (
	contextpkg "context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"

	"github.com/tliron/khutulun/api"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/khutulun/util"
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

	port       int
	grpcServer *grpc.Server
	conductor  *Conductor
}

func NewGRPC(conductor *Conductor) *GRPC {
	return &GRPC{
		port:      8181,
		conductor: conductor,
	}
}

func (self *GRPC) Start() error {
	self.grpcServer = grpc.NewServer()
	api.RegisterConductorServer(self.grpcServer, self)

	if listener, err := net.Listen("tcp", fmt.Sprintf(":%d", self.port)); err == nil {
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
	grpcLog.Info("getVersion")

	return &version, nil
}

// api.ConductorServer interface
func (self *GRPC) ListHosts(empty *emptypb.Empty, server api.Conductor_ListHostsServer) error {
	grpcLog.Info("listHosts")

	if self.conductor.cluster != nil {
		for _, member := range self.conductor.cluster.ListMembers() {
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
	grpcLog.Info("addHost")

	if self.conductor.cluster != nil {
		if err := self.conductor.cluster.AddMembers([]string{identifier.Address}); err == nil {
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
	grpcLog.Info("listNamespaces")

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
func (self *GRPC) ListBundles(listBundles *api.ListBundles, server api.Conductor_ListBundlesServer) error {
	grpcLog.Info("listBundle")

	if identifiers, err := self.conductor.ListBundles(listBundles.Namespace, listBundles.Type.Name); err == nil {
		for _, identifier := range identifiers {
			identifier_ := api.BundleIdentifier{
				Namespace: identifier.Namespace,
				Type:      &api.BundleType{Name: identifier.Type},
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
func (self *GRPC) ListBundleFiles(identifier *api.BundleIdentifier, server api.Conductor_ListBundleFilesServer) error {
	grpcLog.Info("listBundleFiles")

	if bundleFiles, err := self.conductor.ListBundleFiles(identifier.Namespace, identifier.Type.Name, identifier.Name); err == nil {
		for _, bundleFile := range bundleFiles {
			if err := server.Send(&api.BundleFile{
				Path:       bundleFile.Path,
				Executable: bundleFile.Executable,
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
func (self *GRPC) GetBundleFiles(getBundleFiles *api.GetBundleFiles, server api.Conductor_GetBundleFilesServer) error {
	grpcLog.Info("getBundleFiles")

	if lock, err := self.conductor.lockBundle(getBundleFiles.Identifier.Namespace, getBundleFiles.Identifier.Type.Name, getBundleFiles.Identifier.Name, false); err == nil {
		defer func() {
			if err := lock.Unlock(); err != nil {
				grpcLog.Errorf("unlock: %s", err.Error())
			}
		}()

		buffer := make([]byte, BUFFER_SIZE)
		dir := self.conductor.getBundleDir(getBundleFiles.Identifier.Namespace, getBundleFiles.Identifier.Type.Name, getBundleFiles.Identifier.Name)

		for _, path := range getBundleFiles.Paths {
			if file, err := os.Open(filepath.Join(dir, path)); err == nil {
				for {
					if count, err := file.Read(buffer); err == nil {
						content := api.BundleContent{Bytes: buffer[:count]}
						if err := server.Send(&content); err != nil {
							if err := file.Close(); err != nil {
								grpcLog.Errorf("file close: %s", err.Error())
							}
							return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
						}
					} else {
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
func (self *GRPC) SetBundleFiles(server api.Conductor_SetBundleFilesServer) error {
	grpcLog.Info("setBundleFiles")

	if first, err := server.Recv(); err == nil {
		if first.Start != nil {
			namespace := first.Start.Identifier.Namespace
			type_ := first.Start.Identifier.Type.Name
			name := first.Start.Identifier.Name
			if lock, err := self.conductor.lockBundle(namespace, type_, name, true); err == nil {
				defer func() {
					if err := lock.Unlock(); err != nil {
						grpcLog.Errorf("unlock: %s", err.Error())
					}
				}()

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
							// TODO: don't overwrite .lock file
							if file != nil {
								if err := file.Close(); err != nil {
									return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
								}
							}
							path := filepath.Join(self.conductor.getBundleDir(namespace, type_, name), content.File.Path)
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
							return statuspkg.Errorf(codes.Aborted, "message must container \"fileStart\"")
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
								file.Close()
							}
							return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
						}
					}
				}

				if file != nil {
					file.Close()
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
func (self *GRPC) RemoveBundle(context contextpkg.Context, bundleIdentifer *api.BundleIdentifier) (*emptypb.Empty, error) {
	grpcLog.Info("removeBundle")

	if err := self.conductor.DeleteBundle(bundleIdentifer.Namespace, bundleIdentifer.Type.Name, bundleIdentifer.Name); err == nil {
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
	grpcLog.Info("listResources")

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
	grpcLog.Info("interact")

	return util.Interact(server, map[string]util.InteractFunc{
		"host": func(start *api.Interaction_Start) error {
			if len(start.Identifier) != 2 {
				return statuspkg.Errorf(codes.InvalidArgument, "malformed identifier for host: %s", start.Identifier)
			}

			host := start.Identifier[1]

			command := util.NewCommand(start, grpcLog)

			var relay string
			if self.conductor.cluster != nil {
				if self.conductor.cluster.cluster.LocalNode().Name != host {
					relay = fmt.Sprintf("%s:%d", host, 8181)
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
			command := self.conductor.getBundleMainFile("common", "plugin", name)

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
