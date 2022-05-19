package delegate

import (
	"bytes"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/tliron/khutulun/api"
	cloutpkg "github.com/tliron/puccini/clout"
)

func CloutToAPI(clout *cloutpkg.Clout) (*api.Clout, error) {
	if clout != nil {
		if clout_, err := cbor.Marshal(clout); err == nil {
			return &api.Clout{Cbor: clout_}, err
		} else {
			return nil, err
		}
	} else {
		return nil, nil
	}
}

func CloutFromAPI(clout *api.Clout) (*cloutpkg.Clout, error) {
	if clout != nil {
		if clout.Cbor != nil {
			return cloutpkg.Read(bytes.NewReader(clout.Cbor), "cbor")
		} else {
			return cloutpkg.Read(strings.NewReader(clout.Yaml), "yaml")
		}
	} else {
		return nil, nil
	}
}
