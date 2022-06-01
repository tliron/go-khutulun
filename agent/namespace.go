package agent

import (
	"io/ioutil"
	"os"
	"sort"

	"github.com/tliron/kutil/util"
)

func (self *Agent) ListNamespaces() ([]string, error) {
	if files, err := ioutil.ReadDir(self.statePath); err == nil {
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

func (self *Agent) namespaceToNamespaces(namespace string) ([]string, error) {
	if namespace == "" {
		return self.ListNamespaces()
	} else {
		return []string{namespace}, nil
	}
}
