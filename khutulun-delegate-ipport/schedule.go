package main

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/khutulun/delegate"
	"github.com/tliron/khutulun/sdk"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Delegate) Schedule(namespace string, serviceName string, clout *cloutpkg.Clout, coercedClout *cloutpkg.Clout) (*cloutpkg.Clout, []delegate.Next, error) {
	connections, err := sdk.GetCloutConnections(coercedClout)
	if err != nil {
		return nil, nil, err
	}
	if len(connections) == 0 {
		return nil, nil, nil
	}

	state := sdk.NewState("/mnt/khutulun")
	lock, hostEndpoints, err := LockAndGetHostEndpoints(state, self.host)
	if err != nil {
		return nil, nil, err
	}
	defer commonlog.CallAndLogError(lock.Unlock, "unlock", log)

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
		if err := SetHostEndpoints(state, hostEndpoints); err == nil {
			return clout, nil, nil
		} else {
			return nil, nil, err
		}
	} else {
		return nil, nil, nil
	}
}
