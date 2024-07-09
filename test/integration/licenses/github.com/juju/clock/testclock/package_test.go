// Copyright 2022 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testclock_test

import (
	gotesting "testing"
	"time"

	gc "gopkg.in/check.v1"
)

func TestAll(t *gotesting.T) {
	gc.TestingT(t)
}

const (
	shortWait = 50 * time.Millisecond
	longWait  = time.Second
)
