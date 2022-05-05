package client

import (
	"io"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (self *Client) GetVersion() (string, error) {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	if r, err := self.client.GetVersion(context, new(emptypb.Empty)); err == nil {
		return r.Version, nil
	} else {
		return "", err
	}
}

type Host struct {
	Name        string `yaml:"name" json:"name"`
	GRPCAddress string `yaml:"grpcAddress" json:"grpcAddress"`
}

func (self *Client) ListHosts() ([]Host, error) {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	if client, err := self.client.ListHosts(context, new(emptypb.Empty)); err == nil {
		var hosts []Host

		for {
			identifier, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return nil, err
				}
			}

			hosts = append(hosts, Host{
				Name:        identifier.Name,
				GRPCAddress: identifier.GrpcAddress,
			})
		}

		return hosts, nil
	} else {
		return nil, err
	}
}

func (self *Client) AddHost(gossipAddress string) error {
	context, cancel := self.newContextWithTimeout()
	defer cancel()

	if _, err := self.client.AddHost(context, &api.AddHost{GossipAddress: gossipAddress}); err == nil {
		return nil
	} else {
		return util.UnpackGRPCError(err)
	}
}
