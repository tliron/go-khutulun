package util

import (
	"github.com/tliron/khutulun/api"
)

// Both api.Agent_InteractServer and api.Agent_InteractClient
type GRPCInteractor interface {
	Send(*api.Interaction) error
	Recv() (*api.Interaction, error)
}
