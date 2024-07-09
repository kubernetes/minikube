// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package fslock_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gc "gopkg.in/check.v1"

	"github.com/juju/fslock"
)

func Test(t *testing.T) {
	gc.TestingT(t)
}

const (
	shortWait = 10 * time.Millisecond
	longWait  = 10 * shortWait
)

type fslockSuite struct{}

var _ = gc.Suite(&fslockSuite{})

func (s *fslockSuite) TestLockNoContention(c *gc.C) {
	path := filepath.Join(c.MkDir(), "testing")
	lock := fslock.New(path)

	started := make(chan struct{})
	acquired := make(chan struct{})
	go func() {
		close(started)
		err := lock.Lock()
		close(acquired)
		c.Assert(err, gc.IsNil)
	}()

	select {
	case <-started:
		// good, goroutine started.
	case <-time.After(shortWait * 2):
		c.Fatalf("timeout waiting for goroutine to start")
	}

	select {
	case <-acquired:
		// got the lock. good.
	case <-time.After(shortWait * 2):
		c.Fatalf("Timed out waiting for lock acquisition.")
	}

	err := lock.Unlock()
	c.Assert(err, gc.IsNil)
}

func (s *fslockSuite) TestLockBlocks(c *gc.C) {
	path := filepath.Join(c.MkDir(), "testing")
	lock := fslock.New(path)

	kill := make(chan struct{})

	// this will block until the other process has the lock.
	procDone := LockFromAnotherProc(c, path, kill)

	defer func() {
		close(kill)
		// now wait for the other process to exit so the file will be unlocked.
		select {
		case <-procDone:
		case <-time.After(time.Second):
		}
	}()

	started := make(chan struct{})
	acquired := make(chan struct{})
	go func() {
		close(started)
		err := lock.Lock()
		close(acquired)
		lock.Unlock()
		c.Assert(err, gc.IsNil)
	}()

	select {
	case <-started:
		// good, goroutine started.
	case <-time.After(shortWait * 2):
		c.Fatalf("timeout waiting for goroutine to start")
	}

	// Waiting for something not to happen is inherently hard...
	select {
	case <-acquired:
		c.Fatalf("Unexpected lock acquisition")
	case <-time.After(shortWait * 2):
		// all good.
	}
}

func (s *fslockSuite) TestTryLock(c *gc.C) {
	lock := fslock.New(filepath.Join(c.MkDir(), "testing"))

	err := lock.TryLock()
	c.Assert(err, gc.IsNil)
	lock.Unlock()
}

func (s *fslockSuite) TestTryLockNoBlock(c *gc.C) {
	path := filepath.Join(c.MkDir(), "testing")
	lock := fslock.New(path)

	kill := make(chan struct{})

	// this will block until the other process has the lock.
	procDone := LockFromAnotherProc(c, path, kill)

	defer func() {
		close(kill)
		// now wait for the other process to exit so the file will be unlocked.
		select {
		case <-procDone:
		case <-time.After(time.Second):
		}
	}()

	started := make(chan struct{})
	result := make(chan error)
	go func() {
		close(started)
		result <- lock.TryLock()
	}()

	select {
	case <-started:
		// good, goroutine started.
	case <-time.After(shortWait):
		c.Fatalf("timeout waiting for goroutine to start")
	}

	// Wait for trylock to fail.
	select {
	case err := <-result:
		// yes, I know this is redundant with the assert below, but it makes the
		// failed test message clearer.
		if err == nil {
			c.Fatalf("lock succeeded, but should have errored out")
		}
		// This should be the error from trylock failing.
		c.Assert(err, gc.Equals, fslock.ErrLocked)
	case <-time.After(shortWait):
		c.Fatalf("took too long to fail trylock")
	}
}

func (s *fslockSuite) TestUnlockedWithTimeout(c *gc.C) {
	lock := fslock.New(filepath.Join(c.MkDir(), "testing"))

	err := lock.LockWithTimeout(shortWait)
	c.Assert(err, gc.IsNil)
	lock.Unlock()
}

func (s *fslockSuite) TestLockWithTimeout(c *gc.C) {
	path := filepath.Join(c.MkDir(), "testing")
	lock := fslock.New(path)
	defer lock.Unlock()

	kill := make(chan struct{})

	// this will block until the other process has the lock.
	procDone := LockFromAnotherProc(c, path, kill)

	defer func() {
		close(kill)
		// now wait for the other process to exit so the file will be unlocked.
		select {
		case <-procDone:
		case <-time.After(time.Second):
		}
	}()

	started := make(chan struct{})
	result := make(chan error)
	go func() {
		close(started)
		result <- lock.LockWithTimeout(shortWait)
	}()

	select {
	case <-started:
		// good, goroutine started.
	case <-time.After(shortWait * 2):
		c.Fatalf("timeout waiting for goroutine to start")
	}

	// Wait for timeout.
	select {
	case err := <-result:
		// yes, I know this is redundant with the assert below, but it makes the
		// failed test message clearer.
		if err == nil {
			c.Fatalf("lock succeeded, but should have timed out")
		}
		// This should be the error from the lock timing out.
		c.Assert(err, gc.Equals, fslock.ErrTimeout)
	case <-time.After(shortWait * 2):
		c.Fatalf("lock took too long to timeout")
	}
}

func (s *fslockSuite) TestStress(c *gc.C) {
	const lockAttempts = 200
	const concurrentLocks = 10

	var counter = new(int64)
	// Use atomics to update lockState to make sure the lock isn't held by
	// someone else. A value of 1 means locked, 0 means unlocked.
	var lockState = new(int32)

	var wg sync.WaitGroup

	dir := c.MkDir()

	var stress = func(name string) {
		defer wg.Done()
		lock := fslock.New(filepath.Join(dir, "testing"))
		for i := 0; i < lockAttempts; i++ {
			err := lock.Lock()
			c.Assert(err, gc.IsNil)
			state := atomic.AddInt32(lockState, 1)
			c.Assert(state, gc.Equals, int32(1))
			// Tell the go routine scheduler to give a slice to someone else
			// while we have this locked.
			runtime.Gosched()
			// need to decrement prior to unlock to avoid the race of someone
			// else grabbing the lock before we decrement the state.
			atomic.AddInt32(lockState, -1)
			err = lock.Unlock()
			c.Assert(err, gc.IsNil)
			// increment the general counter
			atomic.AddInt64(counter, 1)
		}
	}

	for i := 0; i < concurrentLocks; i++ {
		wg.Add(1)
		go stress(fmt.Sprintf("Lock %d", i))
	}
	wg.Wait()
	c.Assert(*counter, gc.Equals, int64(lockAttempts*concurrentLocks))
}

// LockFromAnotherProc will launch a process and block until that process has
// created the lock file.  If we time out waiting for the other process to take
// the lock, this function will fail the current test.
func LockFromAnotherProc(c *gc.C, path string, kill chan struct{}) (done chan struct{}) {
	cmd := exec.Command(os.Args[0], "-test.run", "TestLockFromOtherProcess")
	cmd.Env = append(
		// We must preserve os.Environ() on Windows,
		// or the subprocess will fail in weird and
		// wonderful ways.
		os.Environ(),
		"FSLOCK_TEST_HELPER_WANTED=1",
		"FSLOCK_TEST_HELPER_PATH="+path,
	)

	if err := cmd.Start(); err != nil {
		c.Fatalf("error starting other proc: %v", err)
	}

	done = make(chan struct{})

	go func() {
		cmd.Wait()
		close(done)
	}()

	go func() {
		select {
		case <-kill:
			// this may fail, but there's not much we can do about it
			_ = cmd.Process.Kill()
		case <-done:
		}
	}()

	for x := 0; x < 10; x++ {
		time.Sleep(shortWait)
		if _, err := os.Stat(path); err == nil {
			// file created by other process, let's continue
			break
		}
		if x == 9 {
			c.Fatalf("timed out waiting for other process to start")
		}
	}
	return done
}

func TestLockFromOtherProcess(t *testing.T) {
	if os.Getenv("FSLOCK_TEST_HELPER_WANTED") == "" {
		return
	}
	filename := os.Getenv("FSLOCK_TEST_HELPER_PATH")
	lock := fslock.New(filename)
	err := lock.Lock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error locking %q: %v", filename, err)
		os.Exit(1)
	}
	time.Sleep(longWait)
	err = lock.Unlock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error unlocking %q: %v", filename, err)
		os.Exit(1)
	}
	os.Exit(0)
}
