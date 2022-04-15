package client

import (
	contextpkg "context"
	"errors"
	"fmt"
	"time"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/configuration"
	"google.golang.org/grpc"
)

const TIMEOUT = 10 * time.Second

//
// Client
//

type Client struct {
	conn    *grpc.ClientConn
	client  api.ConductorClient
	context contextpkg.Context
}

func NewClientFromConfiguration(configurationPath string, clusterName string) (*Client, error) {
	if client, err := configuration.LoadOrNewClient(configurationPath); err == nil {
		if cluster := client.GetCluster(clusterName); cluster != nil {
			target := fmt.Sprintf("%s:%d", cluster.IP, cluster.Port)
			return NewClient(target)
		} else {
			if clusterName == "" {
				return nil, errors.New("no cluster specified")
			} else {
				return nil, fmt.Errorf("unknown cluster: %q", clusterName)
			}
		}
	} else {
		return nil, err
	}
}

func NewClient(target string) (*Client, error) {
	if conn, err := grpc.Dial(target, grpc.WithInsecure()); err == nil {
		return &Client{
			conn:    conn,
			client:  api.NewConductorClient(conn),
			context: contextpkg.Background(),
		}, nil
	} else {
		return nil, err
	}
}

func (self *Client) Close() error {
	return self.conn.Close()
}

func (self *Client) newContextWithTimeout() (contextpkg.Context, contextpkg.CancelFunc) {
	return contextpkg.WithTimeout(self.context, TIMEOUT)
}

func (self *Client) newContextWithCancel() (contextpkg.Context, contextpkg.CancelFunc) {
	return contextpkg.WithCancel(self.context)
}
