package sdk

import (
	"github.com/tliron/go-ard"
)

func ScheduleHost(capability ard.Value, host string) bool {
	if hostAttribute, ok := ard.With(capability).Get("attributes").Get("host").StringMap(); ok {
		if hostValue, ok := hostAttribute["$value"]; ok {
			if hostValue == host {
				return false
			}
		}
		hostAttribute["$value"] = host
		return true
	} else {
		return false
	}
}

func ScheduleIPPort(relationship ard.Value, ip string, port int64) bool {
	attributes := ard.With(relationship).Get("attributes")
	if ip_, ok := attributes.Get("ip").StringMap(); ok {
		ip_["$value"] = ip

		if port_, ok := attributes.Get("port").StringMap(); ok {
			port_["$value"] = port
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}
