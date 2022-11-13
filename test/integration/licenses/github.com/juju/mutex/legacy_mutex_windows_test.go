// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mutex

import (
	"fmt"
	"os"
	"syscall"
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type legacyMutexSuite struct{}

var _ = gc.Suite(&legacyMutexSuite{})

func (s *legacyMutexSuite) spec() Spec {
	// On the off chance there are multiple tests running, we shouldn't
	// clash on the lock name.
	return Spec{
		Name:  fmt.Sprintf("testing-%d", os.Getpid()),
		Clock: &fakeClock{time.Millisecond},
		Delay: 10 * time.Second,
	}
}

func (s *legacyMutexSuite) acquireSemaphore(c *gc.C, name string) syscall.Handle {
	handle, err := s.tryAcquireSemaphore(name)
	c.Assert(err, jc.ErrorIsNil)
	return handle
}

func (s *legacyMutexSuite) tryAcquireSemaphore(name string) (syscall.Handle, error) {
	return createSemaphore("juju-" + name)
}

func (s *legacyMutexSuite) TestLockLegacyContentionTimeout(c *gc.C) {
	spec := s.spec()
	handle := s.acquireSemaphore(c, spec.Name)
	defer syscall.Close(handle)
	defer releaseSemaphore(handle)

	spec.Timeout = time.Nanosecond
	_, err := Acquire(spec)
	c.Assert(err, gc.Equals, ErrTimeout)
}

func (s *legacyMutexSuite) TestLockLegacyContentionRelease(c *gc.C) {
	spec := s.spec()
	handle := s.acquireSemaphore(c, spec.Name)
	defer syscall.Close(handle)
	defer releaseSemaphore(handle)

	type result struct {
		releaser Releaser
		err      error
	}

	acquired := make(chan result)
	go func() {
		r, err := Acquire(spec)
		acquired <- result{r, err}
	}()

	select {
	case <-acquired:
		// mutex shouldn't be acquired until the
		// semaphore is released.
		c.Fatal("unexpected mutex acquisition")
	case <-time.After(50 * time.Millisecond):
	}

	releaseSemaphore(handle)
	select {
	case result := <-acquired:
		c.Assert(result.err, jc.ErrorIsNil)
		defer result.releaser.Release()

		// The mutex should be holding the semaphore open.
		handle, err := s.tryAcquireSemaphore(spec.Name)
		c.Assert(err, gc.Equals, ErrTimeout)
		c.Assert(handle, gc.Equals, syscall.InvalidHandle)

		// Releasing the mutex should release the semaphore.
		result.releaser.Release()
		handle = s.acquireSemaphore(c, spec.Name)
		releaseSemaphore(handle)
		syscall.Close(handle)
	case <-time.After(10 * time.Second):
		c.Fatal("timed out waiting for mutex acquisition")
	}
}

type fakeClock struct {
	delay time.Duration
}

func (f *fakeClock) After(time.Duration) <-chan time.Time {
	return time.After(f.delay)
}

func (f *fakeClock) Now() time.Time {
	return time.Now()
}
