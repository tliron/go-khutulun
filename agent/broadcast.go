package agent

import (
	"errors"
	"fmt"
	"net"

	"github.com/tliron/kutil/format"
	"github.com/tliron/kutil/util"
)

const DEFAULT_MAX_MESSAGE_SIZE = 8192

//
// Broadcaster
//

type Broadcaster struct {
	Protocol string
	address  string
	Port     int

	connection *net.UDPConn
}

func NewBroadcaster(protocol string, address string, port int) *Broadcaster {
	return &Broadcaster{
		Protocol: protocol,
		address:  address,
		Port:     port,
	}
}

func (self *Broadcaster) Start() error {
	if address, err := util.ToBroadcastIPAddress(self.address); err == nil {
		if udpAddr, err := net.ResolveUDPAddr(self.Protocol, util.JoinIPAddressPort(address, self.Port)); err == nil {
			broadcastLog.Noticef("starting broadcaster on %s", udpAddr.String())
			self.connection, err = net.DialUDP(self.Protocol, nil, udpAddr)
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
		length := len(message)
		if n, err := self.connection.Write(message); (err == nil) && (n == length) {
			return nil
		} else if err != nil {
			return err
		} else {
			return fmt.Errorf("write %d bytes instead of %d", n, length)
		}
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
	Protocol       string
	Inter          *net.Interface
	Address        string
	Port           int
	Receive        ReceiveFunc
	MaxMessageSize int
	Ignore         []*net.UDPAddr

	connection *net.UDPConn
}

func NewReceiver(protocol string, inter *net.Interface, address string, port int, receive ReceiveFunc) *Receiver {
	return &Receiver{
		Protocol:       protocol,
		Inter:          inter,
		Address:        address,
		Port:           port,
		Receive:        receive,
		MaxMessageSize: DEFAULT_MAX_MESSAGE_SIZE,
	}
}

func (self *Receiver) Start() error {
	if address, err := net.ResolveUDPAddr(self.Protocol, util.JoinIPAddressPort(self.Address, self.Port)); err == nil {
		broadcastLog.Noticef("starting receiver on %s", address.String())
		if self.connection, err = net.ListenMulticastUDP(self.Protocol, self.Inter, address); err == nil {
			if err := self.connection.SetReadBuffer(self.MaxMessageSize); err != nil {
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
	buffer := make([]byte, self.MaxMessageSize)
	for {
		if count, address, err := self.connection.ReadFromUDP(buffer); err == nil {
			if self.ignore(address) {
				broadcastLog.Debugf("ignoring broadcast from: %s", address.String())
				continue
			}

			message := buffer[:count]
			broadcastLog.Debugf("received broadcast: %s", message)
			self.Receive(address, message)
		} else {
			broadcastLog.Info("receiver closed")
			return
		}
	}
}

func (self *Receiver) ignore(address *net.UDPAddr) bool {
	for _, ignore := range self.Ignore {
		if util.IsUDPAddrEqual(address, ignore) {
			return true
		}
	}
	return false
}
