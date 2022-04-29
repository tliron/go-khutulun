package host

import (
	"errors"
	"os"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/logging"
	problemspkg "github.com/tliron/kutil/problems"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
)

func (self *Host) OpenClout(namespace string, serviceName string) (fslock.Handle, *cloutpkg.Clout, error) {
	if lock, err := self.lockPackage(namespace, "clout", serviceName, false); err == nil {
		cloutPath := self.getPackageMainFile(namespace, "clout", serviceName)
		log.Debugf("reading clout: %q", cloutPath)
		if clout, err := cloutpkg.Load(cloutPath, "yaml"); err == nil {
			return lock, clout, nil
		} else {
			logging.CallAndLogError(lock.Unlock, "unlock", log)
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

func (self *Host) SaveClout(serviceNamespace string, serviceName string, clout *cloutpkg.Clout) error {
	cloutPath := self.getPackageMainFile(serviceNamespace, "clout", serviceName)
	log.Infof("writing to %q", cloutPath)
	if file, err := os.OpenFile(cloutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err == nil {
		defer file.Close()

		return format.WriteYAML(clout, file, "  ", false)
	} else {
		return err
	}
}

func (self *Host) CoerceClout(clout *cloutpkg.Clout) error {
	problems := problemspkg.NewProblems(nil)
	js.Coerce(clout, problems, self.urlContext, true, "yaml", true, false, false)
	if problems.Empty() {
		return nil
	} else {
		return errors.New(problems.String())
	}
}
