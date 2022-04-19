package conductor

import (
	"errors"

	"github.com/danjacques/gofslock/fslock"
	problemspkg "github.com/tliron/kutil/problems"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
)

func (self *Conductor) OpenClout(namespace string, serviceName string) (fslock.Handle, *cloutpkg.Clout, error) {
	if lock, err := self.lockBundle(namespace, "clout", serviceName, false); err == nil {
		cloutPath := self.getBundleMainFile(namespace, "clout", serviceName)
		log.Debugf("reading clout: %q", cloutPath)
		if clout, err := cloutpkg.Load(cloutPath, "yaml"); err == nil {
			return lock, clout, nil
		} else {
			if err := lock.Unlock(); err != nil {
				log.Errorf("unlock: %s", err.Error())
			}
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

func (self *Conductor) GetClout(namespace string, serviceName string, coerce bool) (*cloutpkg.Clout, error) {
	if lock, clout, err := self.OpenClout(namespace, serviceName); err == nil {
		defer func() {
			if err := lock.Unlock(); err != nil {
				grpcLog.Errorf("unlock: %s", err.Error())
			}
		}()

		if coerce {
			if err := self.CoerceClout(clout); err != nil {
				return nil, err
			}
		}

		return clout, nil
	} else {
		return nil, err
	}
}

func (self *Conductor) CoerceClout(clout *cloutpkg.Clout) error {
	problems := problemspkg.NewProblems(nil)
	js.Coerce(clout, problems, self.urlContext, true, "yaml", true, false, false)
	if problems.Empty() {
		return nil
	} else {
		return errors.New(problems.String())
	}
}
