// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

// Package fslock provides a cross-process mutex based on file locks.
//
// It is built on top of flock for linux and darwin, and LockFileEx on Windows.
package fslock

// ErrTimeout indicates that the lock attempt timed out.
var ErrTimeout error = timeoutError("lock timeout exceeded")

type timeoutError string

func (t timeoutError) Error() string {
	return string(t)
}
func (timeoutError) Timeout() bool {
	return true
}

// ErrLocked indicates TryLock failed because the lock was already locked.
var ErrLocked error = trylockError("fslock is already locked")

type trylockError string

func (t trylockError) Error() string {
	return string(t)
}

func (trylockError) Temporary() bool {
	return true
}
