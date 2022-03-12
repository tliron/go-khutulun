package conductor

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/danjacques/gofslock/fslock"
)

type ArtifactIdentifier struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
}

func (self *Conductor) ListArtifacts(namespace string, type_ string) ([]ArtifactIdentifier, error) {
	if namespaces, err := self.namespaceToNamespaces(namespace); err == nil {
		var identifiers []ArtifactIdentifier
		for _, namespace_ := range namespaces {
			if files, err := ioutil.ReadDir(self.getArtifactTypeDir(namespace_, type_)); err == nil {
				for _, file := range files {
					if file.IsDir() {
						identifiers = append(identifiers, ArtifactIdentifier{
							Namespace: namespace_,
							Type:      type_,
							Name:      file.Name(),
						})
					}
				}
			} else {
				if !os.IsNotExist(err) {
					return nil, err
				}
			}
		}
		return identifiers, nil
	} else {
		return nil, err
	}
}

func (self *Conductor) ReadArtifact(namespace string, type_ string, name string) (fslock.Handle, io.ReadCloser, error) {
	if lock, err := self.lockArtifact(namespace, type_, name, false); err == nil {
		path := self.getArtifactFile(namespace, type_, name)
		log.Infof("reading from %q", path)
		if file, err := os.Open(path); err == nil {
			return lock, file, nil
		} else {
			lock.Unlock()
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

func (self *Conductor) WriteArtifact(namespace string, type_ string, name string) (fslock.Handle, io.WriteCloser, error) {
	if lock, err := self.lockArtifact(namespace, type_, name, true); err == nil {
		path := self.getArtifactFile(namespace, type_, name)
		log.Infof("writing %s to %q", type_, path)

		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				lock.Unlock()
				return nil, nil, err
			}
		}

		var mode os.FileMode
		switch type_ {
		case "plugin":
			mode = 0777
		default:
			mode = 0666
		}

		if file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode); err == nil {
			return lock, file, nil
		} else {
			lock.Unlock()
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

func (self *Conductor) DeleteArtifact(namespace string, type_ string, name string) error {
	if lock, err := self.lockArtifact(namespace, type_, name, false); err == nil {
		defer lock.Unlock()
		path := self.getArtifactDir(namespace, type_, name)
		log.Infof("deleting %q", path)
		// TODO: is it OK to delete the lock file while we're holding it?
		return os.RemoveAll(path)
	} else {
		return err
	}
}

func (self *Conductor) getNamespaceDir(namespace string) string {
	if namespace == "" {
		namespace = "_"
	}
	return filepath.Join(self.statePath, namespace)
}

func (self *Conductor) getArtifactTypeDir(namespace string, type_ string) string {
	return filepath.Join(self.getNamespaceDir(namespace), type_)
}

func (self *Conductor) getArtifactDir(namespace string, type_ string, name string) string {
	return filepath.Join(self.getArtifactTypeDir(namespace, type_), name)
}

func (self *Conductor) getArtifactFile(namespace string, type_ string, name string) string {
	switch type_ {
	case "template", "profile", "clout":
		return filepath.Join(self.getArtifactDir(namespace, type_, name), type_+".yaml")
	default:
		return filepath.Join(self.getArtifactDir(namespace, type_, name), name)
	}
}

func (self *Conductor) lockArtifact(namespace string, type_ string, name string, create bool) (fslock.Handle, error) {
	path := filepath.Join(self.getArtifactDir(namespace, type_, name), "lock")
	blocker := newBlocker(time.Second, 5)
	if lock, err := fslock.LockSharedBlocking(path, blocker); err == nil {
		return lock, nil
	} else {
		if os.IsNotExist(err) {
			if create {
				// Touch and try again
				if err := touch(path); err == nil {
					return fslock.LockSharedBlocking(path, blocker)
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
}

func touch(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0777); err == nil {
		if file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err == nil {
			return file.Close()
		} else {
			return err
		}
	} else {
		return err
	}
}

func newBlocker(wait time.Duration, attempts int) fslock.Blocker {
	var attempts_ int
	return func() error {
		time.Sleep(wait)
		if attempts <= 0 {
			return nil
		} else {
			attempts_++
			if attempts_ == attempts {
				return fslock.ErrLockHeld
			} else {
				return nil
			}
		}
	}
}
