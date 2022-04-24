package conductor

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/khutulun/util"
	"github.com/tliron/kutil/logging"
)

const LOCK_FILE = ".lock"

type PackageIdentifier struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
}

type PackageFile struct {
	Path       string
	Executable bool
}

func (self *Conductor) ListPackages(namespace string, type_ string) ([]PackageIdentifier, error) {
	if namespaces, err := self.namespaceToNamespaces(namespace); err == nil {
		var identifiers []PackageIdentifier
		for _, namespace_ := range namespaces {
			if files, err := os.ReadDir(self.getPackageTypeDir(namespace_, type_)); err == nil {
				for _, file := range files {
					name := file.Name()
					if file.IsDir() && !isHidden(name) {
						identifiers = append(identifiers, PackageIdentifier{
							Namespace: namespace_,
							Type:      type_,
							Name:      name,
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

func (self *Conductor) ListPackageFiles(namespace string, type_ string, name string) ([]PackageFile, error) {
	if lock, err := self.lockPackage(namespace, type_, name, false); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", log)

		path := self.getPackageDir(namespace, type_, name)
		length := len(path) + 1
		var files []PackageFile
		if err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
			if !entry.IsDir() {
				if stat, err := os.Stat(path); err == nil {
					files = append(files, PackageFile{
						Path:       path[length:],
						Executable: stat.Mode()&0100 != 0,
					})
				} else {
					return err
				}
			}
			return nil
		}); err == nil {
			return files, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Conductor) ReadPackageFile(namespace string, type_ string, name string, path string) (fslock.Handle, io.ReadCloser, error) {
	if lock, err := self.lockPackage(namespace, type_, name, false); err == nil {
		path = filepath.Join(self.getPackageDir(namespace, type_, name), path)
		log.Debugf("reading from %q", path)
		if file, err := os.Open(path); err == nil {
			return lock, file, nil
		} else {
			logging.CallAndLogError(lock.Unlock, "unlock", log)
			return nil, nil, err
		}
	} else {
		return nil, nil, err
	}
}

func (self *Conductor) DeletePackage(namespace string, type_ string, name string) error {
	if lock, err := self.lockPackage(namespace, type_, name, false); err == nil {
		defer logging.CallAndLogError(lock.Unlock, "unlock", log)

		path := self.getPackageDir(namespace, type_, name)
		log.Infof("deleting package %q", path)
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

func (self *Conductor) getPackageTypeDir(namespace string, type_ string) string {
	return filepath.Join(self.getNamespaceDir(namespace), type_)
}

func (self *Conductor) getPackageDir(namespace string, type_ string, name string) string {
	return filepath.Join(self.getPackageTypeDir(namespace, type_), name)
}

func (self *Conductor) getPackageMainFile(namespace string, type_ string, name string) string {
	dir := self.getPackageDir(namespace, type_, name)
	switch type_ {
	case "template":
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				path := filepath.Join(dir, entry.Name())
				if filepath.Ext(path) == ".yaml" {
					return path
				}
			}
		}
		return ""

	case "plugin":
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				path := filepath.Join(dir, entry.Name())
				if stat, err := os.Stat(path); err == nil {
					if util.IsExecutable(stat.Mode()) {
						return path
					}
				}
			}
		}
		return ""

	case "profile", "clout":
		return filepath.Join(dir, type_+".yaml")

	default:
		return filepath.Join(dir, name)
	}
}

func (self *Conductor) lockPackage(namespace string, type_ string, name string, create bool) (fslock.Handle, error) {
	path := filepath.Join(self.getPackageDir(namespace, type_, name), LOCK_FILE)
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
