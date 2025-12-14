// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package monotonic

import (
	"time"
	_ "unsafe"
)

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64

// Now returns the current time in nanoseconds from a monotonic clock.
//
// The result is guaranteed to not jump due to NTP or other changes to
// system time, which may jump forward or backwards. Instead, in response to
// such changes, the clock frequency is adjusted slowly.
func Now() time.Duration {
	return time.Duration(nanotime())
}
