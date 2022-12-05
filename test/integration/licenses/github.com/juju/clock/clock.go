// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package clock

import "time"

// Clock provides an interface for dealing with clocks.
type Clock interface {
	// Now returns the current clock time.
	Now() time.Time

	// After waits for the duration to elapse and then sends the
	// current time on the returned channel.
	After(time.Duration) <-chan time.Time

	// AfterFunc waits for the duration to elapse and then calls f in its own goroutine.
	// It returns a Timer that can be used to cancel the call using its Stop method.
	AfterFunc(d time.Duration, f func()) Timer

	// NewTimer creates a new Timer that will send the current time
	// on its channel after at least duration d.
	NewTimer(d time.Duration) Timer
}

// Alarm returns a channel that will have the time sent on it at some point
// after the supplied time occurs.
//
// This is short for c.After(t.Sub(c.Now())).
func Alarm(c Clock, t time.Time) <-chan time.Time {
	return c.After(t.Sub(c.Now()))
}

// The Timer type represents a single event.
// A Timer must be created with AfterFunc.
// This interface follows time.Timer's methods but provides easier mocking.
type Timer interface {
	// When the Timer expires, the current time will be sent on the
	// channel returned from Chan, unless the Timer was created by
	// AfterFunc.
	Chan() <-chan time.Time

	// Reset changes the timer to expire after duration d.
	// It returns true if the timer had been active, false if
	// the timer had expired or been stopped.
	Reset(time.Duration) bool

	// Stop prevents the Timer from firing. It returns true if
	// the call stops the timer, false if the timer has already expired or been stopped.
	// Stop does not close the channel, to prevent a read
	// from the channel succeeding incorrectly.
	Stop() bool
}
