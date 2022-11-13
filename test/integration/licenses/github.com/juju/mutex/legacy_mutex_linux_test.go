// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mutex_test

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/mutex"
)

type legacyMutexSuite struct{}

var _ = gc.Suite(&legacyMutexSuite{})

func (s *legacyMutexSuite) spec() mutex.Spec {
	// On the off chance there are multiple tests running, we shouldn't
	// clash on the lock name.
	return mutex.Spec{
		Name:  fmt.Sprintf("testing-%d", os.Getpid()),
		Clock: &fakeClock{time.Millisecond},
		Delay: 10 * time.Second,
	}
}

func (s *legacyMutexSuite) listenUnixSocket(c *gc.C, name string) io.Closer {
	l, err := s.tryListenUnixSocket(name)
	c.Assert(err, jc.ErrorIsNil)
	return l
}

func (s *legacyMutexSuite) tryListenUnixSocket(name string) (io.Closer, error) {
	path := filepath.Join("@/var/lib/juju/mutex-", name)
	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		panic(err)
	}
	return net.ListenUnix("unix", addr)
}

func (s *legacyMutexSuite) TestLockLegacyContentionTimeout(c *gc.C) {
	spec := s.spec()
	l := s.listenUnixSocket(c, spec.Name)
	defer l.Close()

	spec.Timeout = time.Nanosecond
	_, err := mutex.Acquire(spec)
	c.Assert(err, gc.Equals, mutex.ErrTimeout)
}

func (s *legacyMutexSuite) TestLockLegacyContentionRelease(c *gc.C) {
	spec := s.spec()
	l := s.listenUnixSocket(c, spec.Name)
	defer l.Close()

	type result struct {
		releaser mutex.Releaser
		err      error
	}

	acquired := make(chan result)
	go func() {
		r, err := mutex.Acquire(spec)
		acquired <- result{r, err}
	}()

	select {
	case <-acquired:
		// mutex shouldn't be acquired until the
		// unix socket we're holding is closed.
		c.Fatal("unexpected mutex acquisition")
	case <-time.After(50 * time.Millisecond):
	}

	l.Close()
	select {
	case result := <-acquired:
		c.Assert(result.err, jc.ErrorIsNil)
		defer result.releaser.Release()

		// The mutex should be holding the unix socket open.
		_, err := s.tryListenUnixSocket(spec.Name)
		c.Assert(err.(*net.OpError).Err.(*os.SyscallError).Err, gc.Equals, syscall.EADDRINUSE)

		// Releasing the mutex should clsoe the unix socket.
		result.releaser.Release()
		l, err := s.tryListenUnixSocket(spec.Name)
		c.Assert(err, jc.ErrorIsNil)
		l.Close()
	case <-time.After(10 * time.Second):
		c.Fatal("timed out waiting for mutex acquisition")
	}
}
