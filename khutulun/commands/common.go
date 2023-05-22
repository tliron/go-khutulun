package commands

import (
	"github.com/tliron/commonlog"
)

const toolName = "khutulun"

var log = commonlog.GetLogger(toolName)

var clusterName string
var pseudoTerminal bool
var unpack string
