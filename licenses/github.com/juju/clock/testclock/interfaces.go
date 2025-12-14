// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testclock

import (
	"time"

	"github.com/juju/clock"
)

// AdvanceableClock is a clock that can be advanced to trigger timers/trigger timers earlier
// than they would otherwise.
type AdvanceableClock interface {
	clock.Clock
	Advance(time.Duration)
}
