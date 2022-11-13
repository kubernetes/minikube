// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mutex

import (
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/juju/errors"
)

var (
	modkernel32                = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex            = modkernel32.NewProc("CreateMutexW")
	procReleaseMutex           = modkernel32.NewProc("ReleaseMutex")
	procWaitForMultipleObjects = modkernel32.NewProc("WaitForMultipleObjects")
	procCreateEvent            = modkernel32.NewProc("CreateEventW")
	procSetEvent               = modkernel32.NewProc("SetEvent")
)

const (
	_ERROR_ALREADY_EXISTS = 183
	_WAIT_OBJECT_0        = 0x00000000
	_WAIT_ABANDONED_0     = 0x00000080
	_WAIT_TIMEOUT         = 0x00000102
	_INFINITE             = 0xffffffff
)

type mutex struct {
	name     string
	mu       sync.Mutex
	release  chan struct{}
	released chan struct{}
}

func acquire(spec Spec, timeoutCh <-chan time.Time) (Releaser, error) {

	// Avoid further work if the acquisition
	// has already been cancelled.
	select {
	case <-spec.Cancel:
		return nil, ErrCancelled
	default:
	}

	m := &mutex{
		name:     spec.Name,
		release:  make(chan struct{}),
		released: make(chan struct{}),
	}

	releaseCh := m.release
	errCh := make(chan error, 1)
	go func() {
		defer close(m.released)

		// ReleaseMutex must be called
		// from the same OS thread that
		// owns the mutex, i.e. the one
		// that acquires it.
		//
		// Note that this also prevents
		// any other goroutines from
		// running on the same OS thread
		// while the mutex is held, so
		// we prevent reentrant lock
		// acquisitions.
		runtime.LockOSThread()

		handle, err := acquireMutex(
			"juju-mutex-"+spec.Name,
			spec.Timeout,
			spec.Cancel,
		)
		errCh <- err
		if err != nil {
			return
		}

		// Wait until the mutex is Release()d, and
		// then release and close the mutex.
		<-releaseCh

		// NOTE(axw) it is not enough to just close
		// the handle, it must be released first.
		// It will only be abandoned when the OS
		// thread underlying this goroutine exits.
		// See: https://msdn.microsoft.com/en-us/library/windows/desktop/ms684266(v=vs.85).aspx
		if err := releaseMutex(handle); err != nil {
			panic(err)
		}
		if err := syscall.Close(handle); err != nil {
			panic(err)
		}
	}()
	if err := <-errCh; err != nil {
		// The acquireMutex call failed, which means
		// the goroutine will exit. Wait for it to
		// clean up before returning.
		<-m.released
		return nil, err
	}
	return m, nil
}

// Release implements Releaser.
func (m *mutex) Release() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.release != nil {
		close(m.release)
		m.release = nil
	}
	<-m.released
}

func acquireMutex(
	name string,
	timeout time.Duration,
	cancel <-chan struct{},
) (syscall.Handle, error) {
	mutexName, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return syscall.InvalidHandle, errors.Trace(err)
	}

	result, _, errno := syscall.Syscall(
		procCreateMutex.Addr(), 3,
		uintptr(0), // attributes
		uintptr(0), // initial owner
		uintptr(unsafe.Pointer(mutexName)),
	)
	if result == 0 {
		if errno != 0 {
			return syscall.InvalidHandle, errno
		}
		return syscall.InvalidHandle, syscall.EINVAL
	}

	// Mutex was successfully created, or already exists.
	// In either case, we don't own the mutex, and must
	// wait for it below.
	handle := syscall.Handle(result)

	handles := []syscall.Handle{handle}
	if cancel != nil {
		eventHandle, err := createEvent()
		if err != nil {
			return syscall.InvalidHandle, errors.Trace(err)
		}
		defer syscall.Close(eventHandle)

		// Put the event handle first, so
		// cancellation takes precedence.
		handles = []syscall.Handle{eventHandle, handle}
		done := make(chan struct{})
		defer close(done)
		go func() {
			select {
			case <-done:
			case <-cancel:
				setEvent(eventHandle)
			}
		}()
	}

	if timeout == 0 {
		// -1 indicates to waitForMultipleObjects
		// that it should wait indefinitely.
		timeout = -1
	}
	i, err := waitForMultipleObjects(handles, timeout)
	if err != nil {
		syscall.Close(handle)
		return syscall.InvalidHandle, err
	}
	if i+1 == len(handles) {
		return handle, nil
	}
	syscall.Close(handle)
	return syscall.InvalidHandle, ErrCancelled
}

func releaseMutex(handle syscall.Handle) error {
	result, _, errno := syscall.Syscall(procReleaseMutex.Addr(), 3, uintptr(handle), uintptr(1), 0)
	if result == 0 {
		if errno != 0 {
			return errno
		}
		return syscall.EINVAL
	}
	return nil
}

// waitForMultipleObjects calls the WaitForMultipleObjects Windows API, with the
// given handles and timeout. If the timeout value is negative, there is no timeout.
//
// The return value is the index of the handle that was signalled, or an error.
//
// See: https://msdn.microsoft.com/en-us/library/windows/desktop/ms687025(v=vs.85).aspx
func waitForMultipleObjects(handles []syscall.Handle, timeout time.Duration) (int, error) {
	timeoutMillis := uintptr(_INFINITE)
	if timeout >= 0 {
		timeoutMillis = uintptr(timeout.Seconds() * 1000)
	}
	result, _, errno := syscall.Syscall6(
		procWaitForMultipleObjects.Addr(), 4,
		uintptr(len(handles)),
		uintptr(unsafe.Pointer(&handles[0])),
		uintptr(0), // FALSE (don't wait for all)
		timeoutMillis,
		0, 0,
	)
	if result == _WAIT_TIMEOUT {
		return -1, ErrTimeout
	}
	if result >= _WAIT_OBJECT_0 && result < uintptr(_WAIT_OBJECT_0+len(handles)) {
		return int(result - _WAIT_OBJECT_0), nil
	}
	if result >= _WAIT_ABANDONED_0 && result < uintptr(_WAIT_ABANDONED_0+len(handles)) {
		// The thread that owned the mutex exited without releasing the
		// mutex. Ownership of the mutex is still granted as above.
		return int(result - _WAIT_ABANDONED_0), nil
	}
	if errno != 0 {
		return -1, errno
	}
	return -1, syscall.GetLastError()
}

func createEvent() (syscall.Handle, error) {
	result, _, errno := syscall.Syscall6(
		procCreateEvent.Addr(), 4,
		uintptr(0), // lpEventAttributes
		uintptr(1), // bManualReset
		uintptr(0), // bInitialState
		uintptr(0), // lpName
		0, 0,
	)
	if result == 0 {
		if errno != 0 {
			return syscall.Handle(0), errno
		}
		return syscall.Handle(0), syscall.EINVAL
	}
	return syscall.Handle(result), nil
}

func setEvent(handle syscall.Handle) error {
	result, _, errno := syscall.Syscall(procSetEvent.Addr(), 1, uintptr(handle), 0, 0)
	if result == 0 {
		if errno != 0 {
			return errno
		}
		return syscall.EINVAL
	}
	return nil
}
