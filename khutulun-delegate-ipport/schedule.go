package main

import (
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) Schedule(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	connections := sdk.GetCloutConnections(coercedClout)
	if len(connections) == 0 {
		return nil, nil, nil
	}

	state := sdk.NewState("/mnt/khutulun")
	hostEndpoints, err := GetHostEndpoints(state, self.host)
	if err != nil {
		return nil, nil, err
	}

	var changed bool
	for _, connection := range connections {
		if (connection.IP != "") && (connection.Port != 0) {
			continue
		}

		endpoint := Endpoint{
			Namespace: namespace,
			Service:   serviceName,
			Name:      connection.Name,
		}

		if hostEndpoints.Reserve(&endpoint) {
			if edge, err := connection.Find(clout); err == nil {
				if sdk.ScheduleIPPort(edge.Properties, endpoint.Address, endpoint.Port) {
					log.Infof("reserved port %d for %s/%s->%s", endpoint.Port, namespace, serviceName, connection.Name)
					changed = true
				} else {
					// TODO
				}
			} else {
				log.Errorf("%s", err.Error())
			}
		} else {
			log.Warningf("could not reserve port for %s/%s->%s", namespace, serviceName, connection.Name)
		}
	}

	if changed {
		if err := SaveHostEndpoints(state, hostEndpoints); err == nil {
			return clout, nil, nil
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, nil
	}
}

func GetHostEndpoints(state *sdk.State, host string) (*HostEndpoints, error) {
	hostEndpoints := NewHostEndpoints(host, "")
	if err := state.GetHostFileYAML(host, "endpoints.yaml", hostEndpoints); err != nil {
		if host_, err := state.GetHost(host); err == nil {
			hostEndpoints.Address = host_.Address
			hostEndpoints.AddPortRange(9000, 9999)
		} else {
			return nil, err
		}
	}
	return hostEndpoints, nil
}

func SaveHostEndpoints(state *sdk.State, hostEndpoints *HostEndpoints) error {
	return state.SetHostFileYAML(hostEndpoints.Host, "endpoints.yaml", hostEndpoints)
}
