package main

import (
	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/sdk"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

// delegate.Delegate interface
func (self *Delegate) Interact(server sdk.GRPCInteractor, start *api.Interaction_Start) error {
	return statuspkg.Error(codes.Unimplemented, "not implemented")
}
