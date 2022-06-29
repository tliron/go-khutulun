package sdk

import (
	"github.com/tliron/kutil/ard"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"
)

var stateLog = logging.GetLogger("khutulun.state")
var watcherLog = logging.GetLogger("khutulun.watcher")

var ardReflector = ard.NewReflector()

func init() {
	ardReflector.IgnoreMissingStructFields = true
	ardReflector.NilMeansZero = true
	ardReflector.StructFieldNameMapper = util.ToKebabCase
}
