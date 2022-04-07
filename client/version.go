package client

import (
	"google.golang.org/protobuf/types/known/emptypb"
)

func (self *Client) GetVersion() (string, error) {
	context, cancel := self.newContext()
	defer cancel()

	if r, err := self.client.GetVersion(context, new(emptypb.Empty)); err == nil {
		return r.Version, nil
	} else {
		return "", err
	}
}
