// Copyright 2015-2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testclock

import (
	"fmt"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/juju/clock"
	"github.com/juju/errors"
	"github.com/juju/loggo"
)

// timer implements a mock clock.Timer for testing purposes.
type timer struct {
	deadline time.Time
	clock    *Clock
	c        chan time.Time
	// trigger is called when the timer expires. It is
	// called with the clock mutex held and will not block.
	trigger func()
	stack   []byte
}

// Reset is part of the clock.Timer interface.
func (t *timer) Reset(d time.Duration) bool {
	return t.clock.reset(t, d)
}

// Stop is part of the clock.Timer interface.
func (t *timer) Stop() bool {
	return t.clock.stop(t)
}

// Chan is part of the clock.Timer interface.
func (t *timer) Chan() <-chan time.Time {
	return t.c
}

// Clock implements a mock clock.Clock for testing purposes.
type Clock struct {
	mu           sync.Mutex
	now          time.Time
	waiting      []*timer // timers waiting to fire, sorted by deadline.
	notifyAlarms chan struct{}
}

// NewClock returns a new clock set to the supplied time. If your SUT needs to
// call After, AfterFunc, NewTimer or Timer.Reset more than 10000 times: (1)
// you have probably written a bad test; and (2) you'll need to read from the
// Alarms chan to keep the buffer clear.
func NewClock(now time.Time) *Clock {
	return &Clock{
		now:          now,
		notifyAlarms: make(chan struct{}, 10000),
	}
}

// Now is part of the clock.Clock interface.
func (clock *Clock) Now() time.Time {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	return clock.now
}

// After is part of the clock.Clock interface.
func (clock *Clock) After(d time.Duration) <-chan time.Time {
	return clock.NewTimer(d).Chan()
}

func (clock *Clock) NewTimer(d time.Duration) clock.Timer {
	c := make(chan time.Time, 1)
	return clock.addAlarm(d, c, func() {
		c <- clock.now
	})
}

// AfterFunc is part of the clock.Clock interface.
func (clock *Clock) AfterFunc(d time.Duration, f func()) clock.Timer {
	return clock.addAlarm(d, nil, func() {
		go f()
	})
}

func (clock *Clock) addAlarm(d time.Duration, c chan time.Time, trigger func()) *timer {
	defer clock.notifyAlarm()
	clock.mu.Lock()
	defer clock.mu.Unlock()
	t := &timer{
		c:        c,
		deadline: clock.now.Add(d),
		clock:    clock,
		trigger:  trigger,
		stack:    debug.Stack(),
	}
	clock.addTimer(t)
	clock.triggerAll()
	return t
}

// Advance advances the result of Now by the supplied duration, and sends
// the "current" time on all alarms which are no longer "in the future".
func (clock *Clock) Advance(d time.Duration) {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	clock.now = clock.now.Add(d)
	if len(clock.waiting) == 0 {
		loggo.GetLogger("juju.clock").Debugf("advancing a clock that has nothing waiting: cf. https://github.com/juju/juju/wiki/Intermittent-failures")
	}
	clock.triggerAll()
}

// WaitAdvance functions the same as Advance, but only if there is n timers in
// clock.waiting. This came about while fixing lp:1607044 intermittent
// failures.  It turns out that testing.Clock.Advance might advance the time
// and trigger notifications before triggers are set. So we wait a limited time
// 'w' for 'n' timers to show up in clock.waiting, and if they do we advance
// 'd'.
func (clock *Clock) WaitAdvance(d, w time.Duration, n int) error {
	pause := w / 10
	if pause > 10*time.Millisecond {
		pause = 10 * time.Millisecond
	}
	finalTimeout := time.After(w)
	next := time.After(0)
	for {
		select {
		case <-finalTimeout:
			if clock.hasNWaiters(n) {
				clock.Advance(d)
				return nil
			}
			clock.mu.Lock()
			got := len(clock.waiting)
			var stacks string
			for _, t := range clock.waiting {
				stacks += fmt.Sprintf("timer deadline: %v\n%s", t.deadline, string(t.stack))
			}
			clock.mu.Unlock()
			return errors.Errorf(
				"got %d timers added after waiting %s: wanted %d, stacks:\n%s",
				got, w.String(), n, stacks)
		case <-next:
			if clock.hasNWaiters(n) {
				clock.Advance(d)
				return nil
			}
			next = time.After(pause)
		}
	}
}

// hasNWaiters checks if the clock currently has 'n' timers waiting to fire.
func (clock *Clock) hasNWaiters(n int) bool {
	clock.mu.Lock()
	hasWaiters := len(clock.waiting) == n
	clock.mu.Unlock()
	return hasWaiters
}

// Alarms returns a channel on which you can read one value for every call to
// After and AfterFunc; and for every successful Timer.Reset backed by this
// Clock. It might not be elegant but it's necessary when testing time logic
// that runs on a goroutine other than that of the test.
func (clock *Clock) Alarms() <-chan struct{} {
	return clock.notifyAlarms
}

// triggerAll triggers any alarms that are currently due and removes them
// from clock.waiting.
func (clock *Clock) triggerAll() {
	triggered := 0
	for _, t := range clock.waiting {
		if clock.now.Before(t.deadline) {
			break
		}
		t.trigger()
		triggered++
	}
	clock.waiting = clock.waiting[triggered:]
}

// reset is the underlying implementation of clock.Timer.Reset, which may be
// called by any Timer backed by this Clock.
func (clock *Clock) reset(t *timer, d time.Duration) bool {
	defer clock.notifyAlarm()
	clock.mu.Lock()
	defer clock.mu.Unlock()

	found := false
	for _, wt := range clock.waiting {
		if wt == t {
			found = true
		}
	}
	if !found {
		clock.waiting = append(clock.waiting, t)
	}
	t.deadline = clock.now.Add(d)
	sort.Sort(byDeadline(clock.waiting))
	if d <= 0 {
		// If duration is <= 0, that means we should be triggering the
		// Timer right away, as "now" has already occured.
		clock.triggerAll()
	}
	return found
}

// stop is the underlying implementation of clock.Timer.Reset, which may be
// called by any Timer backed by this Clock.
func (clock *Clock) stop(t *timer) bool {
	clock.mu.Lock()
	defer clock.mu.Unlock()

	for i, wt := range clock.waiting {
		if wt == t {
			clock.waiting = removeFromSlice(clock.waiting, i)
			return true
		}
	}
	return false
}

// addTimer adds an alarm at time t.
func (clock *Clock) addTimer(t *timer) {
	clock.waiting = append(clock.waiting, t)
	sort.Sort(byDeadline(clock.waiting))
}

// notifyAlarm sends a value on the channel exposed by Alarms().
func (clock *Clock) notifyAlarm() {
	select {
	case clock.notifyAlarms <- struct{}{}:
	default:
		panic("alarm notification buffer full")
	}
}

// byDeadline is used to sort alarms by time.
type byDeadline []*timer

func (a byDeadline) Len() int           { return len(a) }
func (a byDeadline) Less(i, j int) bool { return a[i].deadline.Before(a[j].deadline) }
func (a byDeadline) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// removeFromSlice removes item at the specified index from the slice.
func removeFromSlice(sl []*timer, index int) []*timer {
	return append(sl[:index], sl[index+1:]...)
}

// AutoAdvancingClock wraps a clock.Clock, calling the Advance
// function whenever After or AfterFunc are called.
type AutoAdvancingClock struct {
	clock.Clock
	Advance func(time.Duration)
}

func (c *AutoAdvancingClock) After(d time.Duration) <-chan time.Time {
	ch := c.Clock.After(d)
	c.Advance(d)
	return ch
}

func (c *AutoAdvancingClock) AfterFunc(d time.Duration, f func()) clock.Timer {
	t := c.Clock.AfterFunc(d, f)
	c.Advance(d)
	return t
}
