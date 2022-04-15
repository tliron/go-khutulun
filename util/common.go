package util

import (
	"github.com/tliron/khutulun/api"
)

type Interactor interface {
	Send(*api.Interaction) error
	Recv() (*api.Interaction, error)
}
