package sdk

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/tliron/kutil/util"
)

func (self *State) GetNamespaceDir(namespace string) string {
	if namespace == "" {
		namespace = "_"
	}
	return filepath.Join(self.RootDir, namespace)
}

func (self *State) ListNamespaces() ([]string, error) {
	if files, err := ioutil.ReadDir(self.RootDir); err == nil {
		var names []string
		for _, file := range files {
			name := file.Name()
			if file.IsDir() && !util.IsFileHidden(name) {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		return names, nil
	} else {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
}

func (self *State) ListNamespacesFor(namespace string) ([]string, error) {
	if namespace == "" {
		return self.ListNamespaces()
	} else {
		return []string{namespace}, nil
	}
}
