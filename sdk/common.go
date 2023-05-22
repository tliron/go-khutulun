package sdk

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/go-ard"
	"github.com/tliron/kutil/util"
)

var stateLog = commonlog.GetLogger("khutulun.state")
var watcherLog = commonlog.GetLogger("khutulun.watcher")

var ardReflector = ard.NewReflector()

func init() {
	ardReflector.IgnoreMissingStructFields = true
	ardReflector.NilMeansZero = true
	ardReflector.StructFieldNameMapper = util.ToKebabCase
}
