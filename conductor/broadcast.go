package conductor

import (
	"errors"
	"net"

	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/util"
)

const DEFAULT_MAX_MESSAGE_SIZE = 8192

//
// Broadcaster
//

type Broadcaster struct {
	protocol   string
	address    *net.UDPAddr
	connection *net.UDPConn
}

func NewBroadcaster(protocol string, address string, port int) (*Broadcaster, error) {
	self := Broadcaster{
		protocol: protocol,
	}
	var err error
	if self.address, err = newUdpAddr(protocol, address, port); err == nil {
		//fmt.Printf("BROADCSTER ZONE: %s\n", self.address.Zone)
		return &self, nil
	} else {
		return nil, err
	}
}

func (self *Broadcaster) Start() error {
	var err error
	self.connection, err = net.DialUDP(self.protocol, nil, self.address)
	return err
}

func (self *Broadcaster) Stop() error {
	if self.connection != nil {
		return self.connection.Close()
	} else {
		return nil
	}
}

func (self *Broadcaster) SendJSON(message any) error {
	if code, err := format.EncodeJSON(message, ""); err == nil {
		return self.Send(util.StringToBytes(code))
	} else {
		return err
	}
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

	protocol       string
	inter          *net.Interface
	address        *net.UDPAddr
	receive        ReceiveFunc
	connection     *net.UDPConn
	maxMessageSize int
}

func NewReceiver(protocol string, inter *net.Interface, address string, port int, receive ReceiveFunc) (*Receiver, error) {
	self := Receiver{
		protocol:       protocol,
		inter:          inter,
		receive:        receive,
		maxMessageSize: DEFAULT_MAX_MESSAGE_SIZE,
	}
	var err error
	if self.address, err = newUdpAddr(protocol, address, port); err == nil {
		//fmt.Printf("RECEIVER ZONE: %s\n", self.address.Zone)
		return &self, nil
	} else {
		return nil, err
	}
}

func (self *Receiver) Start() error {
	var err error
	if self.connection, err = net.ListenMulticastUDP(self.protocol, self.inter, self.address); err == nil {
		self.connection.SetReadBuffer(self.maxMessageSize)

		go func() {
			buffer := make([]byte, self.maxMessageSize)
			for {
				if count, address, err := self.connection.ReadFromUDP(buffer); err == nil {
					if self.ignore(address) {
						clusterLog.Debugf("ignoring broadcast from: %s", address.String())
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
		if address.IP.Equal(ignore_.IP) && (address.Port == ignore_.Port) && (address.Zone == ignore_.Zone) {
			return true
		}
	}
	return false
}
