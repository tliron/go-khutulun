package sdk

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

func JoinAddressPort(address string, port int) string {
	if IsIPv6(address) {
		return fmt.Sprintf("[%s]:%d", address, port)
	} else {
		return fmt.Sprintf("%s:%d", address, port)
	}
}

func NewListener(protocol string, address string, port int) (net.Listener, error) {
	return net.Listen(protocol, JoinAddressPort(address, port))
}

func NewUDPAddr(protocol string, address string, port int) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr(protocol, JoinAddressPort(address, port))
}

func IsUDPAddrEqual(a *net.UDPAddr, b *net.UDPAddr) bool {
	return a.IP.Equal(b.IP) && (a.Port == b.Port) && (a.Zone == b.Zone)
}

func IsIPv6(address string) bool {
	// See: https://stackoverflow.com/questions/22751035/golang-distinguish-ipv4-ipv6
	return strings.Contains(address, ":")
}

func ToReachableAddress(address string) (string, error) {
	if net.ParseIP(address).IsUnspecified() {
		v6 := IsIPv6(address)
		if interfaces, err := net.Interfaces(); err == nil {
			for _, interface_ := range interfaces {
				if (interface_.Flags&net.FlagLoopback == 0) && (interface_.Flags&net.FlagUp != 0) {
					if addrs, err := interface_.Addrs(); err == nil {
						for _, addr := range addrs {
							if addr_, ok := addr.(*net.IPNet); ok {
								//DumpIPInfo(addr_.IP.String())
								if addr_.IP.IsGlobalUnicast() {
									ip := addr_.IP.String()
									if v6 == IsIPv6(ip) {
										return ip, nil
									}
								}
							}
						}
					} else {
						return "", err
					}
				}
			}
			return "", fmt.Errorf("cannot find an equivalent reachable address for: %s", address)
		} else {
			return "", err
		}
	}

	return address, nil
}

func ToBroadcastAddress(address string) (string, error) {
	// Note: net.ParseIP can't parse IPv6 zone
	if ip, err := netip.ParseAddr(address); err == nil {
		if !ip.IsMulticast() {
			return "", fmt.Errorf("not a multicast address: %s", address)
		}

		if IsIPv6(address) && ip.Zone() == "" {
			if interfaces, err := net.Interfaces(); err == nil {
				for _, interface_ := range interfaces {
					//fmt.Printf("%s\n", interface_.Flags.String())
					if (interface_.Flags&net.FlagLoopback == 0) && (interface_.Flags&net.FlagUp != 0) &&
						(interface_.Flags&net.FlagBroadcast != 0) && (interface_.Flags&net.FlagMulticast != 0) {
						return address + "%" + interface_.Name, nil
					}
				}
			} else {
				return "", err
			}
			return "", fmt.Errorf("cannot find a zone for: %s", address)
		}

		return address, nil
	} else {
		return "", err
	}
}

func DumpIPInfo(address string) {
	// Note: net.ParseIP can't parse IPv6 zone
	ip := netip.MustParseAddr(address)
	fmt.Printf("address: %s\n", ip)
	fmt.Printf("  global unicast:            %t\n", ip.IsGlobalUnicast())
	fmt.Printf("  interface local multicast: %t\n", ip.IsInterfaceLocalMulticast())
	fmt.Printf("  link local multicast:      %t\n", ip.IsLinkLocalMulticast())
	fmt.Printf("  link local unicast:        %t\n", ip.IsLinkLocalUnicast())
	fmt.Printf("  loopback:                  %t\n", ip.IsLoopback())
	fmt.Printf("  multicast:                 %t\n", ip.IsMulticast())
	fmt.Printf("  private:                   %t\n", ip.IsPrivate())
	fmt.Printf("  unspecified:               %t\n", ip.IsUnspecified())
}
