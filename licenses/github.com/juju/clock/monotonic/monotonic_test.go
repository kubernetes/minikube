// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package monotonic_test

import (
	"testing"
	"time"

	gc "gopkg.in/check.v1"

	"github.com/juju/clock/monotonic"
)

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

type MonotonicSuite struct {
}

var _ = gc.Suite(&MonotonicSuite{})

func (s *MonotonicSuite) TestNow(c *gc.C) {
	var prev time.Duration
	for i := 0; i < 1000; i++ {
		val := monotonic.Now()
		if val < prev {
			c.Fatal("now is less than previous value")
		}
		prev = val
	}
}
