package sdk

import (
	"fmt"

	"github.com/tliron/go-khutulun/api"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

func GRPCAbortedf(format string, args ...any) error {
	return statuspkg.Errorf(codes.Aborted, format, args...)
}

func GRPCAborted(err error) error {
	return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
}

type grpcStatusError struct {
	status *statuspkg.Status
}

// (error interface)
func (self grpcStatusError) Error() string {
	return fmt.Sprintf("%s: %s", self.status.Code().String(), self.status.Message())
}

func (self grpcStatusError) Unwrap() error {
	return self.status.Err()
}

func UnpackGRPCError(err error) error {
	if status, ok := statuspkg.FromError(err); ok {
		if status.Code() != codes.OK {
			return grpcStatusError{status}
		} else {
			return nil
		}
	} else {
		return err
	}
}

func InteractionErrorDetails(err error) *api.InteractionErrorDetails {
	if statusError_, ok := err.(grpcStatusError); ok {
		for _, details := range statusError_.status.Details() {
			if details_, ok := details.(*api.InteractionErrorDetails); ok {
				return details_
			}
		}
	}
	return nil
}
