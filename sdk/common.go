package sdk

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/util"
)

var ardReflector = ard.NewReflector()

func init() {
	ardReflector.IgnoreMissingStructFields = true
	ardReflector.NilMeansZero = true
	ardReflector.StructFieldNameMapper = util.ToKebabCase
}
