// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mutex

import (
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/juju/errors"
)

var (
	procCreateSemaphore  = modkernel32.NewProc("CreateSemaphoreW")
	procReleaseSemaphore = modkernel32.NewProc("ReleaseSemaphore")
)

type semaphoreMutex struct {
	mu     sync.Mutex
	handle syscall.Handle
}

func acquireLegacy(
	name string,
	clock Clock,
	delay time.Duration,
	timeout <-chan time.Time,
	cancel <-chan struct{},
) (Releaser, error) {
	for {
		impl, err := acquireSemaphore(name)
		if err == nil {
			return impl, nil
		} else if err != ErrTimeout {
			return nil, errors.Trace(err)
		}
		select {
		case <-timeout:
			return nil, ErrTimeout
		case <-cancel:
			return nil, ErrCancelled
		case <-clock.After(delay):
			// no-op, continue and try again
		}
	}
}

func acquireSemaphore(name string) (Releaser, error) {
	handle, err := createSemaphore("juju-" + name)
	if err != nil {
		return nil, err
	}
	return &semaphoreMutex{handle: handle}, nil
}

// Release implements Releaser.
func (m *semaphoreMutex) Release() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.handle == syscall.InvalidHandle {
		return
	}
	if err := releaseSemaphore(m.handle); err != nil {
		panic(err)
	}
	if err := syscall.Close(m.handle); err != nil {
		panic(err)
	}
	m.handle = syscall.InvalidHandle
}

func createSemaphore(name string) (syscall.Handle, error) {
	semName, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return syscall.InvalidHandle, errors.Trace(err)
	}
	result, _, errno := syscall.Syscall6(
		procCreateSemaphore.Addr(), 4,
		uintptr(0), // attributes
		uintptr(0), // initial count
		uintptr(1), // max count
		uintptr(unsafe.Pointer(semName)),
		0, 0,
	)
	if result == 0 {
		if errno != 0 {
			return syscall.InvalidHandle, errno
		}
		return syscall.InvalidHandle, syscall.EINVAL
	}
	handle := syscall.Handle(result)
	if errno == _ERROR_ALREADY_EXISTS {
		handles := []syscall.Handle{handle}
		if _, err := waitForMultipleObjects(handles, 0); err != nil {
			syscall.Close(handle)
			return syscall.InvalidHandle, err
		}
	}
	return handle, nil
}

func releaseSemaphore(handle syscall.Handle) error {
	result, _, errno := syscall.Syscall(procReleaseSemaphore.Addr(), 3, uintptr(handle), uintptr(1), 0)
	if result == 0 {
		if errno != 0 {
			return errno
		}
		return syscall.EINVAL
	}
	return nil
}
