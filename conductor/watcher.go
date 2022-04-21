package conductor

import (
	"github.com/fsnotify/fsnotify"
)

type OnChangedFunc func(path string)

//
// Watcher
//

type Watcher struct {
	watcher   *fsnotify.Watcher
	onChanged OnChangedFunc
}

func NewWatcher(conductor *Conductor, onChanged OnChangedFunc) (*Watcher, error) {
	if watcher, err := fsnotify.NewWatcher(); err == nil {
		if err := watcher.Add(conductor.statePath); err == nil {
			return &Watcher{
				watcher:   watcher,
				onChanged: onChanged,
			}, nil
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

				self.onChanged(event.Name)

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
