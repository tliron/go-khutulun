package conductor

import (
	"io/ioutil"
	"os"
)

func (self *Conductor) ListNamespaces() ([]string, error) {
	if files, err := ioutil.ReadDir(self.statePath); err == nil {
		var names []string
		for _, file := range files {
			if file.IsDir() {
				names = append(names, file.Name())
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

func (self *Conductor) namespaceToNamespaces(namespace string) ([]string, error) {
	if namespace == "" {
		return self.ListNamespaces()
	} else {
		return []string{namespace}, nil
	}
}
