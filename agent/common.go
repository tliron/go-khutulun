package agent

import (
	"github.com/tliron/commonlog"
)

var log = commonlog.GetLogger("khutulun.agent")
var grpcLog = commonlog.NewScopeLogger(log, "grpc")
var gossipLog = commonlog.NewScopeLogger(log, "gossip")
var broadcastLog = commonlog.NewScopeLogger(log, "broadcast")
var httpLog = commonlog.NewScopeLogger(log, "http")
var delegateLog = commonlog.NewScopeLogger(log, "delegate")
