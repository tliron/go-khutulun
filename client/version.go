package client

import (
	"github.com/tliron/khutulun/api"
)

func (self *Client) GetVersion() (string, error) {
	context, cancel := self.newContext()
	defer cancel()

	if r, err := self.client.GetVersion(context, new(api.Empty)); err == nil {
		return r.Version, nil
	} else {
		return "", err
	}
}
