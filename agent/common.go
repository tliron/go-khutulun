package agent

import (
	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("khutulun.agent")
var watcherLog = logging.NewScopeLogger(log, "watcher")
var grpcLog = logging.NewScopeLogger(log, "grpc")
var gossipLog = logging.NewScopeLogger(log, "gossip")
var broadcastLog = logging.NewScopeLogger(log, "broadcast")
var httpLog = logging.NewScopeLogger(log, "http")
var delegateLog = logging.NewScopeLogger(log, "delegate")
