package conductor

import (
	contextpkg "context"
	"fmt"
	"io"
	"net"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/khutulun/api"
	clientpkg "github.com/tliron/khutulun/client"
	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/khutulun/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const BUFFER_SIZE = 4096

var version = api.Version{Version: "0.1.0"}

//
// GRPC
//

type GRPC struct {
	api.UnimplementedConductorServer

	grpcServer *grpc.Server
	conductor  *Conductor
	cluster    *Cluster
}

func NewGRPC(conductor *Conductor, cluster *Cluster) *GRPC {
	return &GRPC{
		conductor: conductor,
		cluster:   cluster,
	}
}

func (self *GRPC) Start() error {
	self.grpcServer = grpc.NewServer()
	api.RegisterConductorServer(self.grpcServer, self)

	if listener, err := net.Listen("tcp", ":8181"); err == nil {
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

	if self.cluster != nil {
		for _, member := range self.cluster.ListMembers() {
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

	if self.cluster != nil {
		if err := self.cluster.AddMembers([]string{identifier.Address}); err == nil {
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
func (self *GRPC) ListArtifacts(listArtifacts *api.ListArtifacts, server api.Conductor_ListArtifactsServer) error {
	grpcLog.Info("listArtifact")

	if identifiers, err := self.conductor.ListArtifacts(listArtifacts.Namespace, listArtifacts.Type.Name); err == nil {
		for _, identifier := range identifiers {
			identifier_ := api.ArtifactIdentifier{
				Namespace: identifier.Namespace,
				Type:      &api.ArtifactType{Name: identifier.Type},
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
func (self *GRPC) GetArtifact(identifier *api.ArtifactIdentifier, server api.Conductor_GetArtifactServer) error {
	grpcLog.Info("getArtifact")

	if lock, reader, err := self.conductor.ReadArtifact(identifier.Namespace, identifier.Type.Name, identifier.Name); err == nil {
		defer lock.Unlock()
		defer reader.Close()
		buffer := make([]byte, BUFFER_SIZE)
		for {
			if count, err := reader.Read(buffer); err == nil {
				content := api.ArtifactContent{Content: &api.ArtifactContent_Bytes{Bytes: buffer[:count]}}
				if err := server.Send(&content); err != nil {
					return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
				}
			} else {
				if err == io.EOF {
					break
				} else {
					return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
				}
			}
		}
	} else {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}

	return nil
}

// api.ConductorServer interface
func (self *GRPC) SetArtifact(server api.Conductor_SetArtifactServer) error {
	grpcLog.Info("setArtifact")

	var namespace string
	var type_ string
	var name string
	var writer io.WriteCloser
	for {
		if content, err := server.Recv(); err == nil {
			switch content_ := content.Content.(type) {
			case *api.ArtifactContent_Identifier:
				namespace = content_.Identifier.Namespace
				type_ = content_.Identifier.Type.Name
				name = content_.Identifier.Name
				var lock fslock.Handle
				var err error
				if lock, writer, err = self.conductor.WriteArtifact(namespace, type_, name); err == nil {
					defer lock.Unlock()
					defer writer.Close()
				} else {
					return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
				}

			case *api.ArtifactContent_Bytes:
				if writer != nil {
					if _, err := writer.Write(content_.Bytes); err != nil {
						return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
					}
				} else {
					return statuspkg.Errorf(codes.InvalidArgument, "first message must be \"identifier\"")
				}
			}
		} else {
			if err == io.EOF {
				break
			} else {
				if writer != nil {
					if err := writer.Close(); err != nil {
						grpcLog.Errorf("close writer: %s", err.Error())
					}
					if err := self.conductor.DeleteArtifact(namespace, type_, name); err != nil {
						grpcLog.Errorf("delete artifact: %s", err.Error())
					}
				}
				return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
			}
		}
	}

	return nil
}

// api.ConductorServer interface
func (self *GRPC) RemoveArtifact(context contextpkg.Context, artifactIdentifer *api.ArtifactIdentifier) (*emptypb.Empty, error) {
	grpcLog.Info("removeArtifact")

	if err := self.conductor.DeleteArtifact(artifactIdentifer.Namespace, artifactIdentifer.Type.Name, artifactIdentifer.Name); err == nil {
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
		"host": func(first *api.Interaction) error {
			if len(first.Start.Identifier) != 2 {
				return statuspkg.Errorf(codes.InvalidArgument, "malformed identifier for host: %s", first.Start.Identifier)
			}

			host := first.Start.Identifier[1]

			command := util.NewCommand(first, grpcLog)

			var relay string
			if self.cluster != nil {
				if self.cluster.cluster.LocalNode().Name != host {
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
				err = client.InteractRelay(server, first)
				grpcLog.Info("interaction ended")
				if err == nil {
					return nil
				} else {
					return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
				}
			}
		},

		"runnable": func(first *api.Interaction) error {
			// TODO: find host for runnable and relay if necessary

			name := "runnable.podman"
			command := self.conductor.getArtifactFile("common", "plugin", name)

			client := plugin.NewRunnableClient(name, command)
			defer client.Close()

			if runnable, err := client.Runnable(); err == nil {
				if err := runnable.Interact(server, first); err == nil {
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
