package util

import (
	"fmt"

	"github.com/tliron/khutulun/api"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

type statusError struct {
	status *statuspkg.Status
}

// error interface
func (self statusError) Error() string {
	return fmt.Sprintf("%s: %s", self.status.Code().String(), self.status.Message())
}

func (self statusError) Unwrap() error {
	return self.status.Err()
}

func UnpackGRPCError(err error) error {
	if status, ok := statuspkg.FromError(err); ok {
		if status.Code() != codes.OK {
			return statusError{status}
		} else {
			return nil
		}
	} else {
		return err
	}
}

func InteractionErrorDetails(err error) *api.InteractionErrorDetails {
	if statusError_, ok := err.(statusError); ok {
		for _, details := range statusError_.status.Details() {
			if details_, ok := details.(*api.InteractionErrorDetails); ok {
				return details_
			}
		}
	}
	return nil
}
