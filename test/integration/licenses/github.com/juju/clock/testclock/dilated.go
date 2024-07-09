// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testclock

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/juju/clock"
)

// NewDilatedWallClock returns a clock that can be sped up or slowed down.
// realSecondDuration is the real duration of a second.
func NewDilatedWallClock(realSecondDuration time.Duration) AdvanceableClock {
	dc := &dilationClock{
		epoch:              time.Now(),
		realSecondDuration: realSecondDuration,
		offsetChanged:      make(chan any),
	}
	dc.offsetChangedCond = sync.NewCond(dc.offsetChangedMutex.RLocker())
	return dc
}

type dilationClock struct {
	epoch              time.Time
	realSecondDuration time.Duration

	// offsetAtomic is the current dilated offset to allow for time jumps/advances.
	offsetAtomic int64
	// offsetChanged is a channel that is closed when timers need to be signaled
	// that there is a offset change coming.
	offsetChanged chan any
	// offsetChangedMutex is a mutex protecting the offsetChanged and is used by
	// the offsetChangedCond.
	offsetChangedMutex sync.RWMutex
	// offsetChangedCond is used to signal timers that they may try to pull the new
	// offset.
	offsetChangedCond *sync.Cond
}

// Now is part of the Clock interface.
func (dc *dilationClock) Now() time.Time {
	dt, _ := dc.nowWithOffset()
	return dt
}

func (dc *dilationClock) nowWithOffset() (time.Time, time.Duration) {
	offset := time.Duration(atomic.LoadInt64(&dc.offsetAtomic))
	realNow := time.Now()
	dt := dilateTime(dc.epoch, realNow, dc.realSecondDuration, offset)
	return dt, offset
}

// After implements Clock.After
func (dc *dilationClock) After(d time.Duration) <-chan time.Time {
	t := newDilatedWallTimer(dc, d, nil)
	return t.c
}

// AfterFunc implements Clock.AfterFunc
func (dc *dilationClock) AfterFunc(d time.Duration, f func()) clock.Timer {
	return newDilatedWallTimer(dc, d, f)
}

// NewTimer implements Clock.NewTimer
func (dc *dilationClock) NewTimer(d time.Duration) clock.Timer {
	return newDilatedWallTimer(dc, d, nil)
}

// Advance implements AdvanceableClock.Advance
func (dc *dilationClock) Advance(d time.Duration) {
	close(dc.offsetChanged)
	dc.offsetChangedMutex.Lock()
	dc.offsetChanged = make(chan any)
	atomic.AddInt64(&dc.offsetAtomic, int64(d))
	dc.offsetChangedCond.Broadcast()
	dc.offsetChangedMutex.Unlock()
}

// dilatedWallTimer implements the Timer interface.
type dilatedWallTimer struct {
	timer      *time.Timer
	dc         *dilationClock
	c          chan time.Time
	target     time.Time
	offset     time.Duration
	after      func()
	done       chan any
	resetChan  chan resetReq
	resetMutex sync.Mutex
	stopChan   chan chan bool
}

type resetReq struct {
	d time.Duration
	r chan bool
}

func newDilatedWallTimer(dc *dilationClock, d time.Duration, after func()) *dilatedWallTimer {
	t := &dilatedWallTimer{
		dc:        dc,
		c:         make(chan time.Time),
		resetChan: make(chan resetReq),
		stopChan:  make(chan chan bool),
	}
	t.start(d, after)
	return t
}

func (t *dilatedWallTimer) start(d time.Duration, after func()) {
	t.dc.offsetChangedMutex.RLock()
	dialatedNow, offset := t.dc.nowWithOffset()
	realDuration := time.Duration(float64(d) * t.dc.realSecondDuration.Seconds())
	t.target = dialatedNow.Add(d)
	t.timer = time.NewTimer(realDuration)
	t.offset = offset
	t.after = after
	t.done = make(chan any)
	go t.run()
}

func (t *dilatedWallTimer) run() {
	defer t.dc.offsetChangedMutex.RUnlock()
	defer close(t.done)
	var sendChan chan time.Time
	var sendTime time.Time
	for {
		select {
		case reset := <-t.resetChan:
			realNow := time.Now()
			dialatedNow := dilateTime(t.dc.epoch, realNow, t.dc.realSecondDuration, t.offset)
			realDuration := time.Duration(float64(reset.d) * t.dc.realSecondDuration.Seconds())
			t.target = dialatedNow.Add(reset.d)
			sendChan = nil
			sendTime = time.Time{}
			reset.r <- t.timer.Reset(realDuration)
		case stop := <-t.stopChan:
			stop <- t.timer.Stop()
			return
		case tt := <-t.timer.C:
			if t.after != nil {
				t.after()
				return
			}
			if sendChan != nil {
				panic("reset should have been called")
			}
			sendChan = t.c
			sendTime = dilateTime(t.dc.epoch, tt, t.dc.realSecondDuration, t.offset)
		case sendChan <- sendTime:
			sendChan = nil
			sendTime = time.Time{}
			return
		case <-t.dc.offsetChanged:
			t.dc.offsetChangedCond.Wait()
			newOffset := time.Duration(atomic.LoadInt64(&t.dc.offsetAtomic))
			if newOffset == t.offset {
				continue
			}
			t.offset = newOffset
			stopped := t.timer.Stop()
			if !stopped {
				continue
			}
			realNow := time.Now()
			dialatedNow := dilateTime(t.dc.epoch, realNow, t.dc.realSecondDuration, t.offset)
			dialatedDuration := t.target.Sub(dialatedNow)
			if dialatedDuration <= 0 {
				sendChan = t.c
				sendTime = dialatedNow
				continue
			}
			realDuration := time.Duration(float64(dialatedDuration) * t.dc.realSecondDuration.Seconds())
			t.timer.Reset(realDuration)
		}
	}
}

// Chan implements Timer.Chan
func (t *dilatedWallTimer) Chan() <-chan time.Time {
	return t.c
}

// Chan implements Timer.Reset
func (t *dilatedWallTimer) Reset(d time.Duration) bool {
	t.resetMutex.Lock()
	defer t.resetMutex.Unlock()
	reset := resetReq{
		d: d,
		r: make(chan bool),
	}
	select {
	case <-t.done:
		t.start(d, nil)
		return true
	case t.resetChan <- reset:
		return <-reset.r
	}
}

// Chan implements Timer.Stop
func (t *dilatedWallTimer) Stop() bool {
	stop := make(chan bool)
	select {
	case <-t.done:
		return false
	case t.stopChan <- stop:
		return <-stop
	}
}

func dilateTime(epoch, realNow time.Time,
	realSecondDuration, dilatedOffset time.Duration) time.Time {
	delta := realNow.Sub(epoch)
	if delta < 0 {
		delta = time.Duration(0)
	}
	return epoch.Add(dilatedOffset).Add(time.Duration(float64(delta) / realSecondDuration.Seconds()))
}
