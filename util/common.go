package util

import (
	"io/fs"

	"github.com/tliron/khutulun/api"
)

type Interactor interface {
	Send(*api.Interaction) error
	Recv() (*api.Interaction, error)
}

func IsExecutable(mode fs.FileMode) bool {
	return mode&0100 != 0
}
