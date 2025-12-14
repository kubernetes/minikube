// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mutex

import (
	"github.com/juju/errors"
)

var (
	ErrTimeout   = errors.New("timeout acquiring mutex")
	ErrCancelled = errors.New("cancelled acquiring mutex")
)
