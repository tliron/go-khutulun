package sdk

import (
	"os"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/transcribe"
	"github.com/tliron/kutil/url"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *State) OpenServiceClout(namespace string, serviceName string, urlContext *url.Context) (fslock.Handle, *cloutpkg.Clout, error) {
	if lock, err := self.LockPackage(namespace, "service", serviceName, false); err == nil {
		cloutPath := self.GetPackageMainFile(namespace, "service", serviceName)
		stateLog.Debugf("reading clout: %q", cloutPath)
		if clout, err := cloutpkg.Load(cloutPath, "yaml", urlContext); err == nil {
			return lock, clout, nil
		} else {
			logging.CallAndLogError(lock.Unlock, "unlock", stateLog)
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
		defer logging.CallAndLogError(file.Close, "file close", stateLog)
		return transcribe.WriteYAML(clout, file, "  ", false)
	} else {
		return err
	}
}
