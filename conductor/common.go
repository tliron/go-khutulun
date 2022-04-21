package conductor

import (
	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("khutulun.conductor")
var watcherLog = logging.GetLogger("khutulun.conductor.watcher")
var grpcLog = logging.GetLogger("khutulun.conductor.grpc")
var clusterLog = logging.GetLogger("khutulun.conductor.cluster")
var httpLog = logging.GetLogger("khutulun.conductor.http")
var reconcileLog = logging.GetLogger("khutulun.conductor.reconcile")
var scheduleLog = logging.GetLogger("khutulun.conductor.schedule")
