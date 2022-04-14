package conductor

import (
	contextpkg "context"
	"errors"
	"io"
	"net"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/khutulun/api"
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
	}
	return nil
}

// api.ConductorServer interface
func (self *GRPC) AddHost(context contextpkg.Context, identifier *api.HostIdentifier) (*emptypb.Empty, error) {
	if self.cluster != nil {
		return new(emptypb.Empty), self.cluster.AddMembers([]string{identifier.Address})
	}
	return new(emptypb.Empty), nil
}

// api.ConductorServer interface
func (self *GRPC) ListNamespaces(empty *emptypb.Empty, server api.Conductor_ListNamespacesServer) error {
	grpcLog.Info("listNamespaces")

	if namespaces, err := self.conductor.ListNamespaces(); err == nil {
		for _, namespace := range namespaces {
			if err := server.Send(&api.Namespace{Name: namespace}); err != nil {
				return err
			}
		}
		return nil
	} else {
		return err
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
				return err
			}
		}
	} else {
		return err
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
					return err
				}
			} else {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}
		}
	} else {
		return err
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
					return err
				}

			case *api.ArtifactContent_Bytes:
				if writer != nil {
					if _, err := writer.Write(content_.Bytes); err != nil {
						return err
					}
				} else {
					// TODO: bytes arrived before an identifier?
				}
			}
		} else {
			if err == io.EOF {
				break
			} else {
				if writer != nil {
					if err := writer.Close(); err != nil {
						grpcLog.Errorf("close writer error: %s", err.Error())
					}
					if err := self.conductor.DeleteArtifact(namespace, type_, name); err != nil {
						grpcLog.Errorf("delete artifact error: %s", err.Error())
					}
				}
				return err
			}
		}
	}

	return nil
}

// api.ConductorServer interface
func (self *GRPC) RemoveArtifact(context contextpkg.Context, artifactIdentifer *api.ArtifactIdentifier) (*emptypb.Empty, error) {
	grpcLog.Info("removeArtifact")
	err := self.conductor.DeleteArtifact(artifactIdentifer.Namespace, artifactIdentifer.Type.Name, artifactIdentifer.Name)
	return new(emptypb.Empty), err
}

// api.ConductorServer interface
func (self *GRPC) DeployService(context contextpkg.Context, deployService *api.DeployService) (*emptypb.Empty, error) {
	grpcLog.Infof("deployService(%q, %q)", deployService.Service.Name, deployService.Template.Name)
	err := self.conductor.DeployService(deployService.Template.Namespace, deployService.Template.Name, deployService.Service.Namespace, deployService.Service.Name)
	return new(emptypb.Empty), err
}

// api.ConductorServer interface
func (self *GRPC) ListResources(listResources *api.ListResources, server api.Conductor_ListResourcesServer) error {
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
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

// api.ConductorServer interface
func (self *GRPC) Interact(server api.Conductor_InteractServer) error {
	grpcLog.Info("interactRun")

	done := make(chan error)
	var kill chan struct{}
	var stdin chan []byte
	var stdout chan []byte
	var stderr chan []byte

	start := func() {
		for {
			select {
			case buffer := <-stdout:
				if buffer == nil {
					grpcLog.Info("stdout closed")
					return
				}
				grpcLog.Debugf("stdout: %q", buffer)
				server.Send(&api.Interaction{
					Stream: "stdout",
					Bytes:  buffer,
				})

			case buffer := <-stderr:
				if buffer == nil {
					grpcLog.Info("stderr closed")
					return
				}
				grpcLog.Debugf("stderr: %q", buffer)
				server.Send(&api.Interaction{
					Stream: "stderr",
					Bytes:  buffer,
				})
			}
		}
	}

	go func() {
		for {
			if interaction, err := server.Recv(); err == nil {
				if stdin == nil {
					if interaction.Start != nil {
						if kill, stdin, stdout, stderr, err = self.conductor.Interact(interaction.Start.Identifier, interaction.Start.PseudoTerminal, done, interaction.Start.Command...); err == nil {
							grpcLog.Info("interaction started")
							go start()
						} else {
							done <- err
							return
						}
					} else {
						done <- errors.New("first message must contain \"start\"")
						return
					}
				}

				switch interaction.Stream {
				case "stdin":
					grpcLog.Debugf("stdin: %q", interaction.Bytes)
					stdin <- interaction.Bytes
				}
			} else {
				if err == io.EOF {
					grpcLog.Info("client closed")
					err = nil
					return
				} else {
					if status, ok := statuspkg.FromError(err); ok {
						if status.Code() == codes.Canceled {
							// We're OK with canceling
							grpcLog.Infof("client canceled")
							err = nil
						}
					}
				}
				kill <- struct{}{}
				done <- err
				return
			}
		}
	}()

	err := <-done
	if stdin != nil {
		close(stdin)
	}
	grpcLog.Info("interaction ended")
	return err
}
