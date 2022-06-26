package sdk

import (
	"github.com/tliron/kutil/ard"
)

func Schedule(capability ard.Value, host string) bool {
	if hostAttribute, ok := ard.NewNode(capability).Get("attributes").Get("host").StringMap(); ok {
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
