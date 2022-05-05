package delegate

import (
	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/util"
)

//
// Delegate
//

type Delegate interface {
	Instantiate(config any) error
	Interact(server util.GRPCInteractor, start *api.Interaction_Start) error
}
