package host

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/tliron/kutil/util"
)

type Change int

const (
	Added   = Change(1)
	Removed = Change(2)
	Changed = Change(3)
)

// fmt.Stringer interface
func (self Change) String() string {
	switch self {
	case Added:
		return "Added"
	case Removed:
		return "Removed"
	case Changed:
		return "Changed"
	default:
		return strconv.Itoa(int(self))
	}
}

type OnChangedFunc func(change Change, identifier []string)

//
// Dir
//

type Dir []string

func NewDir(path string) Dir {
	return Dir(strings.Split(path, string(os.PathSeparator)))
}

// fmt.Stringer interface
func (self Dir) String() string {
	return filepath.Join(self...)
}

func (self Dir) Identifier() ([]string, bool) {
	length := len(self)
	if length == 1 {
		return []string{"namespace", self[0]}, false
	} else if length > 2 {
		return []string{self[1], self[0], self[2]}, length > 3
	} else {
		return nil, false
	}
}

func (self Dir) Equals(dir Dir) bool {
	if len(self) != len(dir) {
		return false
	}
	for index, segment := range self {
		if segment != dir[index] {
			return false
		}
	}
	return true
}

//
// Watcher
//

type Watcher struct {
	host      *Host
	watcher   *fsnotify.Watcher
	onChanged OnChangedFunc
	dirs      []Dir
	lock      util.RWLocker
}

func NewWatcher(host *Host, onChanged OnChangedFunc) (*Watcher, error) {
	var self = Watcher{
		host:      host,
		onChanged: onChanged,
		lock:      util.NewDefaultRWLocker(),
	}

	var err error
	if self.watcher, err = fsnotify.NewWatcher(); err == nil {
		if err := self.sync(); err == nil {
			self.sync()
			self.sync()
			return &self, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *Watcher) Start() {
	watcherLog.Notice("starting watcher")
	go func() {
		for {
			select {
			case event, ok := <-self.watcher.Events:
				if !ok {
					watcherLog.Info("closed watcher")
					return
				}

				// Ignore hidden files
				if isHidden(event.Name) {
					continue
				}

				watcherLog.Debugf("%s %s", event.Op.String(), event.Name)

				switch event.Op {
				case fsnotify.Create:
					if stat, err := os.Stat(event.Name); err == nil {
						if stat.IsDir() {
							dir := self.toDir(event.Name)
							if err := self.add(dir); err != nil {
								watcherLog.Warningf("%s", err.Error())
							}

							if identifier, packageFile := dir.Identifier(); identifier != nil {
								if packageFile {
									self.onChanged(Changed, identifier)
								} else {
									self.onChanged(Added, identifier)
								}
							}
						}
					} else {
						watcherLog.Warningf("%s", err.Error())
					}

				case fsnotify.Remove, fsnotify.Rename:
					// Note: we may receive this twice, once from the dir itself and once from its parent
					if err := self.remove(self.toDir(event.Name)); err != nil {
						watcherLog.Warningf("%s", err.Error())
					}

					dir := self.toDir(event.Name)
					if identifier, packageFile := dir.Identifier(); identifier != nil {
						if packageFile {
							self.onChanged(Changed, identifier)
						} else {
							self.onChanged(Removed, identifier)
						}
					}

				case fsnotify.Write, fsnotify.Chmod:
					dir := self.toDir(event.Name)
					if identifier, packageFile := dir.Identifier(); identifier != nil {
						if packageFile {
							self.onChanged(Changed, identifier)
						}
					}
				}

			case err, ok := <-self.watcher.Errors:
				if !ok {
					watcherLog.Info("closed watcher")
					return
				}

				watcherLog.Errorf("watcher: %s", err.Error())
			}
		}
	}()
}

func (self *Watcher) Stop() error {
	return self.watcher.Close()
}

func (self *Watcher) sync() error {
	self.lock.Lock()
	defer self.lock.Unlock()

	return filepath.WalkDir(self.host.statePath, func(path string, entry fs.DirEntry, err error) error {
		if isHidden(path) {
			return fs.SkipDir
		}
		if entry.IsDir() {
			if err := self.watcher.Add(path); err == nil {
				dir := NewDir(path)
				watcherLog.Debugf("adding dir: %s", dir.String())
				self.dirs = append(self.dirs, dir)
			} else {
				return err
			}
		}
		return nil
	})
}

func (self *Watcher) add(dir Dir) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	watcherLog.Debugf("adding dir: %s", dir.String())
	self.dirs = append(self.dirs, dir)
	if err := self.watcher.Add(self.toPath(dir)); err == nil {
		return nil
	} else {
		return err
	}
}

func (self *Watcher) remove(dir Dir) error {
	self.lock.Lock()
	defer self.lock.Unlock()

	// TODO: improve this
	var err error
	var dirs []Dir
	for _, dir_ := range self.dirs {
		if dir_.Equals(dir) {
			watcherLog.Debugf("removing dir: %s", dir.String())
			self.watcher.Remove(self.toPath(dir)) // we are ignoring errors
		} else {
			dirs = append(dirs, dir_)
		}
	}
	self.dirs = dirs
	return err
}

func (self *Watcher) toDir(path string) Dir {
	statePath := self.host.statePath
	length := len(statePath) + 1
	if path == statePath {
		path = ""
	} else {
		path = path[length:]
	}
	return NewDir(path)
}

func (self *Watcher) toPath(dir Dir) string {
	segments := append([]string{self.host.statePath}, dir...)
	return filepath.Join(segments...)
}
