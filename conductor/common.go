package conductor

import (
	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("khutulun.server")
var grpcLog = logging.GetLogger("khutulun.server.grpc")
var clusterLog = logging.GetLogger("khutulun.server.cluster")
var httpLog = logging.GetLogger("khutulun.server.http")
var reconcileLog = logging.GetLogger("khutulun.server.reconcile")
var scheduleLog = logging.GetLogger("khutulun.server.schedule")
