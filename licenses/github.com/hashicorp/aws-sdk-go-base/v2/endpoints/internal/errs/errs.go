// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	"errors"
)

// IsA indicates whether an error matches an error type.
func IsA[T error](err error) bool {
	_, ok := As[T](err)
	return ok
}

// As is equivalent to errors.As(), but returns the value in-line.
func As[T error](err error) (T, bool) {
	var as T
	ok := errors.As(err, &as)
	return as, ok
}
