package sdk

import (
	contextpkg "context"
	"os"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/commonlog"
	"github.com/tliron/exturl"
	"github.com/tliron/kutil/transcribe"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *State) OpenServiceClout(context contextpkg.Context, namespace string, serviceName string, urlContext *exturl.Context) (fslock.Handle, *cloutpkg.Clout, error) {
	if lock, err := self.LockPackage(namespace, "service", serviceName, false); err == nil {
		cloutPath := self.GetPackageMainFile(namespace, "service", serviceName)
		stateLog.Debugf("reading clout: %q", cloutPath)
		if clout, err := cloutpkg.Load(context, cloutPath, "yaml", urlContext); err == nil {
			return lock, clout, nil
		} else {
			commonlog.CallAndLogError(lock.Unlock, "unlock", stateLog)
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

func (self *State) SaveServiceClout(serviceNamespace string, serviceName string, clout *cloutpkg.Clout) error {
	cloutPath := self.GetPackageMainFile(serviceNamespace, "service", serviceName)
	stateLog.Infof("writing to %q", cloutPath)
	if file, err := os.OpenFile(cloutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err == nil {
		defer commonlog.CallAndLogError(file.Close, "file close", stateLog)
		return transcribe.WriteYAML(clout, file, "  ", false)
	} else {
		return err
	}
}
