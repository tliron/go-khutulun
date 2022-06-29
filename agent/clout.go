package agent

import (
	"io"
	"os"
	"strings"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/kutil/logging"
	problemspkg "github.com/tliron/kutil/problems"
	"github.com/tliron/kutil/transcribe"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
)

func (self *Agent) OpenServiceClout(namespace string, serviceName string) (fslock.Handle, *cloutpkg.Clout, error) {
	if lock, err := self.state.LockPackage(namespace, "service", serviceName, false); err == nil {
		cloutPath := self.state.GetPackageMainFile(namespace, "service", serviceName)
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

func (self *Agent) SaveServiceClout(serviceNamespace string, serviceName string, clout *cloutpkg.Clout) error {
	cloutPath := self.state.GetPackageMainFile(serviceNamespace, "service", serviceName)
	log.Infof("writing to %q", cloutPath)
	if file, err := os.OpenFile(cloutPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err == nil {
		defer logging.CallAndLogError(file.Close, "file close", log)
		return transcribe.WriteYAML(clout, file, "  ", false)
	} else {
		return err
	}
}

func (self *Agent) CoerceClout(clout *cloutpkg.Clout, copy_ bool) (*cloutpkg.Clout, error) {
	coercedClout := clout
	if copy_ {
		var err error
		if coercedClout, err = clout.Copy(); err != nil {
			return nil, err
		}
	}
	problems := problemspkg.NewProblems(nil)
	js.Coerce(coercedClout, problems, self.urlContext, true, "yaml", false, true)
	return coercedClout, problems.ToError(true)
}

func (self *Agent) OpenFile(path string, coerceClout bool) (io.ReadCloser, error) {
	if coerceClout {
		if file, err := os.Open(path); err == nil {
			defer logging.CallAndLogError(file.Close, "file close", log)
			if clout, err := cloutpkg.Read(file, "yaml"); err == nil {
				if clout, err = self.CoerceClout(clout, false); err == nil {
					if code, err := transcribe.EncodeYAML(clout, "  ", false); err == nil {
						return io.NopCloser(strings.NewReader(code)), nil
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return os.Open(path)
	}
}
