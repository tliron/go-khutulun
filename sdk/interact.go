package sdk

import (
	"io"

	"github.com/tliron/commonlog"
	"github.com/tliron/khutulun/api"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

// Both api.Agent_InteractServer and api.Agent_InteractClient
type GRPCInteractor interface {
	Send(*api.Interaction) error
	Recv() (*api.Interaction, error)
}

type InteractFunc func(start *api.Interaction_Start) error

func Interact(server GRPCInteractor, interact map[string]InteractFunc) error {
	if first, err := server.Recv(); err == nil {
		if first.Start != nil {
			if len(first.Start.Identifier) == 0 {
				return statuspkg.Error(codes.InvalidArgument, "no identifier")
			}
			type_ := first.Start.Identifier[0]

			if interact_, ok := interact[type_]; ok {
				return interact_(first.Start)
			} else {
				return statuspkg.Errorf(codes.InvalidArgument, "malformed identifier: %s", first.Start.Identifier)
			}
		} else {
			return statuspkg.Error(codes.InvalidArgument, "first message must contain \"start\"")
		}
	} else {
		return GRPCAborted(err)
	}
}

func InteractRelay(server GRPCInteractor, client GRPCInteractor, start *api.Interaction_Start, log commonlog.Logger) error {
	if err := client.Send(&api.Interaction{Start: start}); err != nil {
		return err
	}

	go func() {
		for {
			if interaction, err := server.Recv(); err == nil {
				if err := client.Send(interaction); err != nil {
					log.Errorf("client send: %s", err.Error())
					return
				}
			} else {
				if err == io.EOF {
					log.Info("client closed")
				} else if statuspkg.Code(err) == codes.Canceled {
					// We're OK with canceling
					log.Infof("client canceled")
				} else {
					log.Errorf("client receive: %s", err.Error())
				}
				return
			}
		}
	}()

	for {
		if interaction, err := client.Recv(); err == nil {
			if err := server.Send(interaction); err != nil {
				return err
			}
		} else {
			if err == io.EOF {
				log.Info("server closed")
				return nil
			} else {
				return err
			}
		}
	}
}
