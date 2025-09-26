package delegate

import (
	"bytes"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/tliron/go-khutulun/api"
	"github.com/tliron/go-kutil/util"
	cloutpkg "github.com/tliron/go-puccini/clout"
	yamlpkg "gopkg.in/yaml.v3"
)

var yaml_ bool = true

func CloutToAPI(clout *cloutpkg.Clout, yaml bool) (*api.Clout, error) {
	if clout != nil {
		if yaml {
			if clout_, err := yamlpkg.Marshal(clout); err == nil {
				return &api.Clout{Yaml: util.BytesToString(clout_)}, err
			} else {
				return nil, err
			}
		} else {
			if clout_, err := cbor.Marshal(clout); err == nil {
				return &api.Clout{Cbor: clout_}, err
			} else {
				return nil, err
			}
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
