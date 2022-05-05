package agent

import (
	"errors"
	"net"

	khutulunutil "github.com/tliron/khutulun/util"
	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/util"
)

const DEFAULT_MAX_MESSAGE_SIZE = 8192

//
// Broadcaster
//

type Broadcaster struct {
	protocol   string
	address    string
	port       int
	connection *net.UDPConn
}

func NewBroadcaster(protocol string, address string, port int) *Broadcaster {
	return &Broadcaster{
		protocol: protocol,
		address:  address,
		port:     port,
	}
}

func (self *Broadcaster) Start() error {
	if address, err := khutulunutil.ToBroadcastAddress(self.address); err == nil {
		if udpAddr, err := khutulunutil.NewUDPAddr(self.protocol, address, self.port); err == nil {
			broadcastLog.Noticef("starting broadcaster on: %s", udpAddr.String())
			self.connection, err = net.DialUDP(self.protocol, nil, udpAddr)
			return err
		} else {
			return err
		}
	} else {
		return err
	}
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
		broadcastLog.Debugf("sending broadcast: %s", message)
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
	address        string
	port           int
	receive        ReceiveFunc
	connection     *net.UDPConn
	maxMessageSize int
}

func NewReceiver(protocol string, inter *net.Interface, address string, port int, receive ReceiveFunc) *Receiver {
	return &Receiver{
		protocol:       protocol,
		inter:          inter,
		address:        address,
		port:           port,
		receive:        receive,
		maxMessageSize: DEFAULT_MAX_MESSAGE_SIZE,
	}
}

func (self *Receiver) Start() error {
	if address, err := khutulunutil.NewUDPAddr(self.protocol, self.address, self.port); err == nil {
		broadcastLog.Noticef("starting receiver on: %s", address.String())
		if self.connection, err = net.ListenMulticastUDP(self.protocol, self.inter, address); err == nil {
			if err := self.connection.SetReadBuffer(self.maxMessageSize); err != nil {
				return err
			}
			go self.read()
			return nil
		} else {
			return err
		}
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

func (self *Receiver) read() {
	buffer := make([]byte, self.maxMessageSize)
	for {
		if count, address, err := self.connection.ReadFromUDP(buffer); err == nil {
			if self.ignore(address) {
				broadcastLog.Debugf("ignoring broadcast from: %s", address.String())
				continue
			}

			message := buffer[:count]
			broadcastLog.Debugf("received broadcast: %s", message)
			self.receive(address, message)
		} else {
			broadcastLog.Info("receiver closed")
			return
		}
	}
}

func (self *Receiver) ignore(address *net.UDPAddr) bool {
	for _, ignore := range self.Ignore {
		if khutulunutil.IsUDPAddrEqual(address, ignore) {
			return true
		}
	}
	return false
}
