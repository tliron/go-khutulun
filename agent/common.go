package agent

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/tliron/kutil/logging"
)

var log = logging.GetLogger("khutulun.agent")
var watcherLog = logging.GetLogger("khutulun.watcher")
var grpcLog = logging.GetLogger("khutulun.grpc")
var gossipLog = logging.GetLogger("khutulun.gossip")
var httpLog = logging.GetLogger("khutulun.http")
var reconcileLog = logging.GetLogger("khutulun.reconcile")
var scheduleLog = logging.GetLogger("khutulun.schedule")

func isHidden(path string) bool {
	return strings.HasPrefix(filepath.Base(path), ".")
}

func newListener(protocol string, address string, port int) (net.Listener, error) {
	return net.Listen(protocol, fmt.Sprintf("[%s]:%d", address, port))
}

func newUdpAddr(protocol string, address string, port int) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr(protocol, fmt.Sprintf("[%s]:%d", address, port))
}

func toReachableAddress(address string) (string, error) {
	if net.ParseIP(address).IsUnspecified() {
		// Find first compatible address
		v6 := strings.Contains(address, ":") // see: https://stackoverflow.com/questions/22751035/golang-distinguish-ipv4-ipv6
		if interfaces, err := net.Interfaces(); err == nil {
			for _, interface_ := range interfaces {
				if (interface_.Flags&net.FlagLoopback == 0) && (interface_.Flags&net.FlagUp != 0) {
					if addrs, err := interface_.Addrs(); err == nil {
						for _, addr := range addrs {
							if addr_, ok := addr.(*net.IPNet); ok {
								ip := addr_.IP.String()
								v6_ := strings.Contains(ip, ":")
								if v6 == v6_ {
									return ip, nil
								}
							}
						}
					} else {
						return "", err
					}
				}
			}
		} else {
			return "", err
		}
	}

	return address, nil
}
