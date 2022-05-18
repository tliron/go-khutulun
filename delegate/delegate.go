package delegate

import (
	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

//
// Delegate
//

type Delegate interface {
	ProcessService(namespace string, serviceName string, phase string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, error)
	//Instantiate(config any) error
	Interact(server sdk.GRPCInteractor, start *api.Interaction_Start) error
}
