package agent

import (
	"io/ioutil"
	"os"
)

func (self *Agent) ListNamespaces() ([]string, error) {
	if files, err := ioutil.ReadDir(self.statePath); err == nil {
		var names []string
		for _, file := range files {
			name := file.Name()
			if file.IsDir() && !isHidden(name) {
				names = append(names, name)
			}
		}
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
