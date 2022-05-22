package delegate

import "github.com/tliron/khutulun/api"

//
// Next
//

type Next struct {
	Host        string
	Phase       string
	Namespace   string
	ServiceName string
}

func (self Next) Equals(next Next) bool {
	return (self.Host == next.Host) && (self.Phase == next.Phase) && (self.Namespace == next.Namespace) && (self.ServiceName == next.ServiceName)
}

func AppendNext(next []Next, host string, phase string, namespace string, serviceName string) []Next {
	next_ := Next{
		Host:        host,
		Phase:       phase,
		Namespace:   namespace,
		ServiceName: serviceName,
	}
	for _, next__ := range next {
		if next_.Equals(next__) {
			return next
		}
	}
	return append(next, next_)
}

func MergeNexts(to []Next, from []Next) []Next {
	for _, next := range from {
		var exists bool
		for _, next_ := range to {
			if next_.Equals(next) {
				exists = true
				break
			}
		}
		if !exists {
			to = append(to, next)
		}
	}
	return to
}

func NextToAPI(next Next) *api.NextService {
	return &api.NextService{
		Host:  next.Host,
		Phase: next.Phase,
		Service: &api.ServiceIdentifier{
			Namespace: next.Namespace,
			Name:      next.ServiceName,
		},
	}
}

func NextsToAPI(next []Next) []*api.NextService {
	if length := len(next); length > 0 {
		next_ := make([]*api.NextService, length)
		for index, next__ := range next {
			next_[index] = NextToAPI(next__)
		}
		return next_
	} else {
		return nil
	}
}

func NextFromAPI(next *api.NextService) Next {
	return Next{
		Host:        next.Host,
		Phase:       next.Phase,
		Namespace:   next.Service.Namespace,
		ServiceName: next.Service.Name,
	}
}

func NextsFromAPI(next []*api.NextService) []Next {
	if length := len(next); length > 0 {
		next_ := make([]Next, length)
		for index, next__ := range next {
			next_[index] = NextFromAPI(next__)
		}
		return next_
	} else {
		return nil
	}
}
