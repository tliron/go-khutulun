package agent

import (
	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("khutulun.agent")
var watcherLog = logging.GetLogger("khutulun.watcher")
var grpcLog = logging.GetLogger("khutulun.grpc")
var gossipLog = logging.GetLogger("khutulun.gossip")
var broadcastLog = logging.GetLogger("khutulun.broadcast")
var httpLog = logging.GetLogger("khutulun.http")
var reconcileLog = logging.GetLogger("khutulun.reconcile")
var scheduleLog = logging.GetLogger("khutulun.schedule")
