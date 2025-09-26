package sdk

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/commonlog"
	"github.com/tliron/go-kutil/util"
)

const LOCK_FILE = ".lock"

func (self *State) GetPackageTypeDir(namespace string, type_ string) string {
	return filepath.Join(self.GetNamespaceDir(namespace), type_)
}

func (self *State) GetPackageDir(namespace string, type_ string, name string) string {
	return filepath.Join(self.GetPackageTypeDir(namespace, type_), name)
}

func (self *State) GetPackageMainFile(namespace string, type_ string, name string) string {
	dir := self.GetPackageDir(namespace, type_, name)
	switch type_ {
	case "service":
		return filepath.Join(dir, "clout.yaml")

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

	case "profile":
		return filepath.Join(dir, "profile.yaml")

	case "delegate":
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				path := filepath.Join(dir, entry.Name())
				if stat, err := os.Stat(path); err == nil {
					if util.IsFileExecutable(stat.Mode()) {
						return path
					}
				}
			}
		}
		return ""

	case "host":
		return filepath.Join(dir, "host.yaml")

	default:
		return filepath.Join(dir, name)
	}
}

func (self *State) ListPackages(namespace string, type_ string) (PackageIdentifiers, error) {
	if namespaces, err := self.ListNamespacesFor(namespace); err == nil {
		var identifiers PackageIdentifiers
		for _, namespace_ := range namespaces {
			if files, err := os.ReadDir(self.GetPackageTypeDir(namespace_, type_)); err == nil {
				for _, file := range files {
					name := file.Name()
					if file.IsDir() && !util.IsFileHidden(name) {
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
		sort.Sort(identifiers)
		return identifiers, nil
	} else {
		return nil, err
	}
}

func (self *State) LockPackage(namespace string, type_ string, name string, create bool) (fslock.Handle, error) {
	path := filepath.Join(self.GetPackageDir(namespace, type_, name), LOCK_FILE)
	blocker := newFsLockBlocker(time.Second, 5)
	if lock, err := fslock.LockSharedBlocking(path, blocker); err == nil {
		return lock, nil
	} else if os.IsNotExist(err) {
		if create {
			// Touch and try again
			if err := util.Touch(path, 0666, 0777); err == nil {
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

func (self *State) ListPackageFiles(namespace string, type_ string, name string) ([]PackageFile, error) {
	if lock, err := self.LockPackage(namespace, type_, name, false); err == nil {
		defer commonlog.CallAndLogError(lock.Unlock, "unlock", stateLog)

		path := self.GetPackageDir(namespace, type_, name)
		length := len(path) + 1
		var files []PackageFile
		if err := filepath.WalkDir(path, func(path string, dirEntry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !dirEntry.IsDir() {
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

func (self *State) OpenPackageFile(namespace string, type_ string, name string, path string) (io.ReadCloser, error) {
	path = filepath.Join(self.GetPackageDir(namespace, type_, name), path)
	stateLog.Debugf("reading from %q", path)
	return os.Open(path)
}

func (self *State) CreatePackageFile(namespace string, type_ string, name string, path string) (io.WriteCloser, error) {
	path = filepath.Join(self.GetPackageDir(namespace, type_, name), path)
	stateLog.Debugf("writing to %q", path)
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
}

func (self *State) LockAndOpenPackageFile(namespace string, type_ string, name string, path string) (*LockedReadCloser, error) {
	if lock, err := self.LockPackage(namespace, type_, name, false); err == nil {
		if reader, err := self.OpenPackageFile(namespace, type_, name, path); err == nil {
			return &LockedReadCloser{reader, lock}, nil
		} else {
			commonlog.CallAndLogError(lock.Unlock, "unlock", stateLog)
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *State) LockAndCreatePackageFile(namespace string, type_ string, name string, path string) (*LockedWriteCloser, error) {
	if lock, err := self.LockPackage(namespace, type_, name, true); err == nil {
		if writer, err := self.CreatePackageFile(namespace, type_, name, path); err == nil {
			return &LockedWriteCloser{writer, lock}, nil
		} else {
			commonlog.CallAndLogError(lock.Unlock, "unlock", stateLog)
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *State) DeletePackage(namespace string, type_ string, name string) error {
	if lock, err := self.LockPackage(namespace, type_, name, false); err == nil {
		defer commonlog.CallAndLogError(lock.Unlock, "unlock", stateLog)

		path := self.GetPackageDir(namespace, type_, name)
		stateLog.Infof("deleting package %q", path)
		if entries, err := os.ReadDir(path); err == nil {
			for _, entry := range entries {
				name := entry.Name()
				if name == LOCK_FILE {
					continue
				}
				name = filepath.Join(path, name)
				if err := os.RemoveAll(name); err != nil {
					return err
				}
			}
		}
		return nil
	} else {
		return err
	}
}

// Utils

func newFsLockBlocker(wait time.Duration, maxAttempts int) fslock.Blocker {
	var attempts int
	return func() error {
		time.Sleep(wait)
		if maxAttempts <= 0 {
			return nil
		} else {
			attempts++
			if attempts == maxAttempts {
				return fslock.ErrLockHeld
			} else {
				return nil
			}
		}
	}
}

//
// PackageIdentifier
//

type PackageIdentifier struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Type      string `json:"type" yaml:"type"`
	Name      string `json:"name" yaml:"name"`
}

//
// PackageIdentifiers
//

type PackageIdentifiers []PackageIdentifier

// sort.Interface interface
func (self PackageIdentifiers) Len() int {
	return len(self)
}

// sort.Interface interface
func (self PackageIdentifiers) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

// sort.Interface interface
func (self PackageIdentifiers) Less(i, j int) bool {
	if c := strings.Compare(self[i].Namespace, self[j].Namespace); c == 0 {
		if c := strings.Compare(self[i].Type, self[j].Type); c == 0 {
			return strings.Compare(self[i].Name, self[j].Name) == -1
		} else {
			return c == 1
		}
	} else {
		return c == -1
	}
}

//
// PackageFile
//

type PackageFile struct {
	Path       string
	Executable bool
}
