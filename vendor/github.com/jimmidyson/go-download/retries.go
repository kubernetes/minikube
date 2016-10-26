package download

import (
	"time"

	multierror "github.com/hashicorp/go-multierror"
)

type retriableError struct {
	err error
}

func (e *retriableError) Error() string {
	return e.err.Error()
}

func retryAfter(attempts int, callback func() error, d time.Duration) error {
	var res *multierror.Error
	if attempts == -1 {
		attempts = ^int(0)
	}
	for i := 0; i < attempts; i++ {
		err := callback()
		if err == nil {
			return nil
		}
		res = multierror.Append(res, err)
		if _, ok := err.(*retriableError); !ok {
			return res
		}
		<-time.After(d)
	}
	return res.ErrorOrNil()
}
