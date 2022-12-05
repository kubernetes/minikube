// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mutex

import (
	"regexp"
	"time"

	"github.com/juju/errors"
)

var (
	validName = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9-]*$")
)

// Releaser defines the Release method that is the only thing that can be done
// to a acquired mutex.
type Releaser interface {
	// Release releases the mutex. Release may be called multiple times, but
	// only the first call will release this instance of the mutex. Release is
	// unable to release the mutex successfully it will call panic to forcibly
	// release the mutex.
	Release()
}

// Clock provides an interface for dealing with clocks.
type Clock interface {
	// After waits for the duration to elapse and then sends the
	// current time on the returned channel.
	After(time.Duration) <-chan time.Time

	// Now returns the current clock time.
	Now() time.Time
}

// Spec defines the name of the mutex and behaviour of the Acquire function.
type Spec struct {
	// Name is required, and must start with a letter and contain at most
	// 40 letters, numbers or dashes.
	Name string

	// Clock must be provided and is exposed for testing purposes.
	Clock Clock

	// Delay defines how often to check for lock acquisition, for
	// compatibility code that requires polling.
	Delay time.Duration

	// Timeout allows the caller to specify how long to wait. If Timeout
	// is zero, then the call will block forever.
	Timeout time.Duration

	// Cancel if signalled will cause the Acquire method to return with ErrCancelled.
	Cancel <-chan struct{}
}

// Acquire will attempt to acquire the named mutex. If the Timout value
// is hit, ErrTimeout is returned. If the Cancel channel is signalled,
// ErrCancelled is returned.
func Acquire(spec Spec) (Releaser, error) {
	if err := spec.Validate(); err != nil {
		return nil, errors.Trace(err)
	}

	var timeout <-chan time.Time
	if spec.Timeout > 0 {
		timeout = spec.Clock.After(spec.Timeout)
	}

	return acquire(spec, timeout)
}

// Validate checks the attributes of Spec for validity.
func (s *Spec) Validate() error {
	if len(s.Name) > 40 {
		return errors.NotValidf("Name longer than 40 characters")
	}
	if !validName.MatchString(s.Name) {
		return errors.NotValidf("Name %q", s.Name)
	}
	if s.Clock == nil {
		return errors.NotValidf("missing Clock")
	}
	if s.Delay <= 0 {
		return errors.NotValidf("non positive Delay")
	}
	if s.Timeout < 0 {
		return errors.NotValidf("negative Timeout")
	}
	return nil
}
