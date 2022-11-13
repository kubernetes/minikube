// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// +build !linux,!windows

package mutex

import (
	"time"
)

func acquireLegacy(
	name string,
	clock Clock,
	delay time.Duration,
	timeout <-chan time.Time,
	cancel <-chan struct{},
) (Releaser, error) {
	return nopReleaser{}, nil
}

type nopReleaser struct{}

// Release is part of the Releaser interface.
func (nopReleaser) Release() {}
