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

	// At waits for the time to pass and then sends the
	// current time on the returned channel.
	At(t time.Time) <-chan time.Time

	// AtFunc waits for the time to pass and then calls f in its own goroutine.
	// It returns an Alarm that can be used to cancel the call using its Stop method.
	AtFunc(t time.Time, f func()) Alarm

	// NewAlarm creates a new Alarm that will send the current time
	// on its channel at or after time t has passed.
	NewAlarm(t time.Time) Alarm
}

// Timer type represents a single event.
// Timers must be created with AfterFunc or NewTimer.
// This interface follows time.Timer's methods but provides easier mocking.
type Timer interface {
	// When the timer expires, the current time will be sent on the
	// channel returned from Chan, unless the timer was created by
	// AfterFunc.
	Chan() <-chan time.Time

	// Reset changes the timer to expire after duration d.
	// It returns true if the timer had been active, false if
	// the timer had expired or been stopped.
	Reset(d time.Duration) bool

	// Stop prevents the Timer from firing. It returns true if
	// the call stops the timer, false if the timer has already expired or been stopped.
	// Stop does not close the channel, to prevent a read
	// from the channel succeeding incorrectly.
	Stop() bool
}

// Alarm type represents a single event.
// Alarms must be created with AtFunc or NewAlarm.
type Alarm interface {
	// When the alarm expires, the current time will be sent on the
	// channel returned from Chan, unless the alarm was created by
	// AtFunc.
	Chan() <-chan time.Time

	// Reset changes the alarm to expire at or after time t.
	// It returns true if the alarm had been active, false if
	// the alarm had fired or been stopped.
	Reset(t time.Time) bool

	// Stop prevents the alarm from firing. It returns true if
	// the call stops the alarm, false if the alarm has already fired or been stopped.
	// Stop does not close the channel, to prevent a read
	// from the channel succeeding incorrectly.
	Stop() bool
}
