package util

import (
	"fmt"

	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

type statusError struct {
	status *statuspkg.Status
}

func (self statusError) Error() string {
	return fmt.Sprintf("%s: %s", self.status.Code().String(), self.status.Message())
}

func (self statusError) Unwrap() error {
	return self.status.Err()
}

func UnpackGrpcError(err error) error {
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
