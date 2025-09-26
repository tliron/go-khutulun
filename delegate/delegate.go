package delegate

import (
	"github.com/tliron/go-khutulun/api"
	"github.com/tliron/go-khutulun/sdk"
	cloutpkg "github.com/tliron/go-puccini/clout"
)

//
// Delegate
//

type Delegate interface {
	ListResources(namespace string, serviceName string, coercedClout *cloutpkg.Clout) ([]Resource, error)
	ProcessService(namespace string, serviceName string, phase string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []Next, error)
	Interact(server sdk.GRPCInteractor, start *api.Interaction_Start) error
}

type Resource struct {
	Type string
	Name string
	Host string
}
