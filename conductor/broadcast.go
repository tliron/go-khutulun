package conductor

import (
	"errors"
	"net"

	"github.com/tliron/kutil/util"
)

const DEFAULT_MAX_MESSAGE_SIZE = 8192

//
// Broadcaster
//

type Broadcaster struct {
	network    string
	address    *net.UDPAddr
	connection *net.UDPConn
}

func NewBroadcaster(network string, address string) (*Broadcaster, error) {
	self := Broadcaster{
		network: network,
	}
	var err error
	if self.address, err = net.ResolveUDPAddr(network, address); err == nil {
		return &self, nil
	} else {
		return nil, err
	}
}

func (self *Broadcaster) Start() error {
	var err error
	self.connection, err = net.DialUDP(self.network, nil, self.address)
	return err
}

func (self *Broadcaster) Stop() error {
	if self.connection != nil {
		return self.connection.Close()
	} else {
		return nil
	}
}

func (self *Broadcaster) SendString(message string) error {
	return self.Send(util.StringToBytes(message))
}

func (self *Broadcaster) Send(message []byte) error {
	if self.connection != nil {
		clusterLog.Infof("sending broadcast: %s", message)
		_, err := self.connection.Write(message)
		return err
	} else {
		return errors.New("not started")
	}
}

func (self *Broadcaster) Address() *net.UDPAddr {
	if self.connection != nil {
		if address, ok := self.connection.LocalAddr().(*net.UDPAddr); ok {
			return address
		} else {
			return nil
		}
	} else {
		return nil
	}
}

//
// Receiver
//

type ReceiveFunc func(address *net.UDPAddr, message []byte)

type Receiver struct {
	Ignore []*net.UDPAddr

	network        string
	address        *net.UDPAddr
	receive        ReceiveFunc
	connection     *net.UDPConn
	maxMessageSize int
}

func NewReceiver(network string, address string, receive ReceiveFunc) (*Receiver, error) {
	self := Receiver{
		network:        network,
		receive:        receive,
		maxMessageSize: DEFAULT_MAX_MESSAGE_SIZE,
	}
	var err error
	if self.address, err = net.ResolveUDPAddr(network, address); err == nil {
		return &self, nil
	} else {
		return nil, err
	}
}

func (self *Receiver) Start() error {
	var err error
	if self.connection, err = net.ListenMulticastUDP(self.network, nil, self.address); err == nil {
		self.connection.SetReadBuffer(self.maxMessageSize)

		go func() {
			buffer := make([]byte, self.maxMessageSize)
			for {
				if count, address, err := self.connection.ReadFromUDP(buffer); err == nil {
					if self.ignore(address) {
						clusterLog.Infof("ignoring broadcast from: %s", address.String())
						continue
					}

					message := buffer[:count]
					clusterLog.Infof("received broadcast: %s", message)
					self.receive(address, message)
				} else {
					clusterLog.Info("receiver closed")
					return
				}
			}
		}()

		return nil
	} else {
		return err
	}
}

func (self *Receiver) Stop() error {
	if self.connection != nil {
		return self.connection.Close()
	} else {
		return nil
	}
}

func (self *Receiver) ignore(address *net.UDPAddr) bool {
	for _, ignore_ := range self.Ignore {
		if address.IP.Equal(ignore_.IP) && (address.Port == ignore_.Port) {
			return true
		}
	}
	return false
}
