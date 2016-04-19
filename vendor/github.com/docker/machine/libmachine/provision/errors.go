package provision

import (
	"errors"
	"fmt"
)

var (
	ErrDetectionFailed = errors.New("OS type not recognized")
)

type ErrDaemonAvailable struct {
	wrappedErr error
}

func (e ErrDaemonAvailable) Error() string {
	return fmt.Sprintf("Unable to verify the Docker daemon is listening: %s", e.wrappedErr)
}

func NewErrDaemonAvailable(err error) ErrDaemonAvailable {
	return ErrDaemonAvailable{
		wrappedErr: err,
	}
}
