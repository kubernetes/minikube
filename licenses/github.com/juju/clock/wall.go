// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package clock

import (
	"time"
)

// WallClock exposes wall-clock time via the Clock interface.
var WallClock wallClock

// ensure that WallClock does actually implement the Clock interface.
var _ Clock = WallClock

// WallClock exposes wall-clock time as returned by time.Now.
type wallClock struct{}

// Now is part of the Clock interface.
func (wallClock) Now() time.Time {
	return time.Now()
}

// After implements Clock.After.
func (wallClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// AfterFunc implements Clock.AfterFunc.
func (wallClock) AfterFunc(d time.Duration, f func()) Timer {
	return wallTimer{time.AfterFunc(d, f)}
}

// NewTimer implements Clock.NewTimer.
func (wallClock) NewTimer(d time.Duration) Timer {
	return wallTimer{time.NewTimer(d)}
}

// wallTimer implements the Timer interface.
type wallTimer struct {
	*time.Timer
}

// Chan implements Timer.Chan.
func (t wallTimer) Chan() <-chan time.Time {
	return t.C
}

// At implements Clock.At.
func (wallClock) At(t time.Time) <-chan time.Time {
	return time.After(time.Until(t))
}

// AtFunc implements Clock.AtFunc.
func (wallClock) AtFunc(t time.Time, f func()) Alarm {
	return wallAlarm{time.AfterFunc(time.Until(t), f)}
}

// NewAlarm implements Clock.NewAlarm.
func (wallClock) NewAlarm(t time.Time) Alarm {
	return wallAlarm{time.NewTimer(time.Until(t))}
}

// wallAlarm implements the Alarm interface.
type wallAlarm struct {
	*time.Timer
}

// Chan implements Alarm.Chan.
func (a wallAlarm) Chan() <-chan time.Time {
	return a.C
}

// Reset implements Alarm.Reset
func (a wallAlarm) Reset(t time.Time) bool {
	return a.Timer.Reset(time.Until(t))
}
