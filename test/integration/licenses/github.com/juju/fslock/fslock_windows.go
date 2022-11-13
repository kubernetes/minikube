// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fslock

import (
	"log"
	"syscall"
	"time"
	"unsafe"
)

var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procCreateEventW = modkernel32.NewProc("CreateEventW")
)

const (
	lockfileExclusiveLock = 2
	fileFlagNormal        = 0x00000080
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Ldate)
}

// Lock implements cross-process locks using syscalls.
// This implementation is based on LockFileEx syscall.
type Lock struct {
	filename string
	handle   syscall.Handle
}

// New returns a new lock around the given file.
func New(filename string) *Lock {
	return &Lock{filename: filename}
}

// TryLock attempts to lock the lock.  This method will return ErrLocked
// immediately if the lock cannot be acquired.
func (l *Lock) TryLock() error {
	err := l.LockWithTimeout(0)
	if err == ErrTimeout {
		// in our case, timing out immediately just means it was already locked.
		return ErrLocked
	}
	return err
}

// Lock locks the lock.  This call will block until the lock is available.
func (l *Lock) Lock() error {
	return l.LockWithTimeout(-1)
}

// Unlock unlocks the lock.
func (l *Lock) Unlock() error {
	return syscall.Close(l.handle)
}

// LockWithTimeout tries to lock the lock until the timeout expires.  If the
// timeout expires, this method will return ErrTimeout.
func (l *Lock) LockWithTimeout(timeout time.Duration) (err error) {
	name, err := syscall.UTF16PtrFromString(l.filename)
	if err != nil {
		return err
	}

	// Open for asynchronous I/O so that we can timeout waiting for the lock.
	// Also open shared so that other processes can open the file (but will
	// still need to lock it).
	handle, err := syscall.CreateFile(
		name,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ,
		nil,
		syscall.OPEN_ALWAYS,
		syscall.FILE_FLAG_OVERLAPPED|fileFlagNormal,
		0)
	if err != nil {
		return err
	}
	l.handle = handle
	defer func() {
		if err != nil {
			syscall.Close(handle)
		}
	}()

	millis := uint32(syscall.INFINITE)
	if timeout >= 0 {
		millis = uint32(timeout.Nanoseconds() / 1000000)
	}

	ol, err := newOverlapped()
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(ol.HEvent)
	err = lockFileEx(handle, lockfileExclusiveLock, 0, 1, 0, ol)
	if err == nil {
		return nil
	}

	// ERROR_IO_PENDING is expected when we're waiting on an asychronous event
	// to occur.
	if err != syscall.ERROR_IO_PENDING {
		return err
	}
	s, err := syscall.WaitForSingleObject(ol.HEvent, millis)

	switch s {
	case syscall.WAIT_OBJECT_0:
		// success!
		return nil
	case syscall.WAIT_TIMEOUT:
		return ErrTimeout
	default:
		return err
	}
}

// newOverlapped creates a structure used to track asynchronous
// I/O requests that have been issued.
func newOverlapped() (*syscall.Overlapped, error) {
	event, err := createEvent(nil, true, false, nil)
	if err != nil {
		return nil, err
	}
	return &syscall.Overlapped{HEvent: event}, nil
}

func lockFileEx(h syscall.Handle, flags, reserved, locklow, lockhigh uint32, ol *syscall.Overlapped) (err error) {
	r1, _, e1 := syscall.Syscall6(procLockFileEx.Addr(), 6, uintptr(h), uintptr(flags), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func createEvent(sa *syscall.SecurityAttributes, manualReset bool, initialState bool, name *uint16) (handle syscall.Handle, err error) {
	var _p0 uint32
	if manualReset {
		_p0 = 1
	}
	var _p1 uint32
	if initialState {
		_p1 = 1
	}

	r0, _, e1 := syscall.Syscall6(procCreateEventW.Addr(), 4, uintptr(unsafe.Pointer(sa)), uintptr(_p0), uintptr(_p1), uintptr(unsafe.Pointer(name)), 0, 0)
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
