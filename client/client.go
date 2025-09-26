package client

import (
	contextpkg "context"
	"errors"
	"fmt"
	"time"

	"github.com/tliron/go-khutulun/api"
	"github.com/tliron/go-khutulun/configuration"
	"github.com/tliron/go-kutil/util"
	"google.golang.org/grpc"
)

const DEFAULT_TIMEOUT = 10 * time.Second

//
// Client
//

type Client struct {
	Timeout time.Duration

	conn    *grpc.ClientConn
	client  api.AgentClient
	context contextpkg.Context
}

func NewClientFromConfiguration(configurationPath string, clusterName string) (*Client, error) {
	if client, err := configuration.LoadOrNewClient(configurationPath); err == nil {
		if cluster := client.GetCluster(clusterName); cluster != nil {
			target := util.JoinIPAddressPort(cluster.IP, cluster.Port)
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
			Timeout: DEFAULT_TIMEOUT,
			conn:    conn,
			client:  api.NewAgentClient(conn),
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
	return contextpkg.WithTimeout(self.context, self.Timeout)
}

func (self *Client) newContextWithCancel() (contextpkg.Context, contextpkg.CancelFunc) {
	return contextpkg.WithCancel(self.context)
}
