// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mutex_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	jt "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/mutex/v2"
)

const (
	shortWait = 10 * time.Millisecond
	longWait  = 10 * shortWait
	appWait   = 5 * time.Second
)

type mutexSuite struct{}

var _ = gc.Suite(&mutexSuite{})

func (s *mutexSuite) TestSpecValidity(c *gc.C) {
	for i, test := range []struct {
		spec mutex.Spec
		err  string
	}{{
		spec: mutex.Spec{Name: "", Clock: &fakeClock{}, Delay: time.Millisecond},
		err:  `Name "" not valid`,
	}, {
		spec: mutex.Spec{Name: "42", Clock: &fakeClock{}, Delay: time.Millisecond},
		err:  `Name "42" not valid`,
	}, {
		spec: mutex.Spec{Name: "a", Clock: &fakeClock{}, Delay: time.Millisecond},
	}, {
		spec: mutex.Spec{Name: "a very very long name that is over the length limit", Clock: &fakeClock{}, Delay: time.Millisecond},
		err:  `Name longer than 40 characters not valid`,
	}, {
		spec: mutex.Spec{Name: "test-42", Clock: &fakeClock{}, Delay: time.Millisecond},
	}, {
		spec: mutex.Spec{Name: "with a space", Clock: &fakeClock{}, Delay: time.Millisecond},
		err:  `Name "with a space" not valid`,
	}, {
		spec: mutex.Spec{Name: "test-42", Delay: time.Millisecond},
		err:  `missing Clock not valid`,
	}, {
		spec: mutex.Spec{Name: "test-42", Clock: &fakeClock{}},
		err:  `non positive Delay not valid`,
	}, {
		spec: mutex.Spec{Name: "test-42", Clock: &fakeClock{}, Delay: -time.Second},
		err:  `non positive Delay not valid`,
	}, {
		spec: mutex.Spec{Name: "test-42", Clock: &fakeClock{}, Delay: time.Millisecond, Timeout: -time.Second},
		err:  `negative Timeout not valid`,
	}} {
		c.Logf("test %d", i+1)
		err := test.spec.Validate()
		if test.err == "" {
			c.Check(err, jc.ErrorIsNil)
		} else {
			c.Check(err, gc.ErrorMatches, test.err)
		}
	}
}

func (s *mutexSuite) spec() mutex.Spec {
	// On the off chance there are multiple tests running, we shouldn't
	// clash on the lock name.
	return mutex.Spec{
		Name:  fmt.Sprintf("testing-%d", os.Getpid()),
		Clock: &fakeClock{time.Millisecond},
		Delay: 10 * time.Second,
	}
}

func (s *mutexSuite) TestLockNoContention(c *gc.C) {
	r, err := mutex.Acquire(s.spec())
	c.Assert(err, jc.ErrorIsNil)
	r.Release()
}

func (s *mutexSuite) TestLockContentionWithinProcessTimeout(c *gc.C) {
	spec := s.spec()
	spec.Timeout = time.Millisecond

	r, err := mutex.Acquire(spec)
	c.Assert(err, jc.ErrorIsNil)
	defer r.Release()

	second, err := mutex.Acquire(spec)
	c.Assert(second, gc.IsNil)
	c.Assert(err, gc.Equals, mutex.ErrTimeout)
}

func (s *mutexSuite) TestLockAfterTimeout(c *gc.C) {
	spec := s.spec()
	spec.Timeout = time.Nanosecond

	r, err := mutex.Acquire(spec)
	c.Assert(err, jc.ErrorIsNil)
	defer r.Release()

	_, err = mutex.Acquire(spec)
	c.Assert(err, gc.Equals, mutex.ErrTimeout)

	r.Release()

	r, err = mutex.Acquire(spec)
	c.Assert(err, jc.ErrorIsNil)
	r.Release()
}

func (s *mutexSuite) TestLockContentionWithinProcessCancel(c *gc.C) {
	cancel := make(chan struct{})
	done := make(chan struct{})

	spec := s.spec()
	spec.Cancel = cancel

	r, err := mutex.Acquire(spec)
	c.Assert(err, jc.ErrorIsNil)
	defer r.Release()

	go func() {
		second, err := mutex.Acquire(spec)
		c.Check(second, gc.IsNil)
		c.Check(err, gc.Equals, mutex.ErrCancelled)
		close(done)
	}()

	close(cancel)
	<-done
}

func (s *mutexSuite) TestSecondReleaseFine(c *gc.C) {
	r, err := mutex.Acquire(s.spec())
	c.Assert(err, jc.ErrorIsNil)
	r.Release()
	r.Release()
}

func (s *mutexSuite) TestDifferentNamesDontBlock(c *gc.C) {
	spec1 := s.spec()
	spec2 := spec1
	spec2.Name = fmt.Sprintf("other-%d", os.Getpid())

	r1, err := mutex.Acquire(spec1)
	c.Assert(err, jc.ErrorIsNil)
	defer r1.Release()

	r2, err := mutex.Acquire(spec2)
	c.Assert(err, jc.ErrorIsNil)
	defer r2.Release()
}

func (s *mutexSuite) TestLockBlocks(c *gc.C) {
	kill := make(chan struct{})
	spec := s.spec()
	// this will block until the other process has the lock.
	remote := LockFromAnotherProc(c, spec.Name, kill, true, true)

	acquired := make(chan struct{})
	done := make(chan struct{})
	go func() {
		r, err := mutex.Acquire(spec)
		if c.Check(err, gc.IsNil) {
			close(acquired)
			r.Release()
		}
		close(done)
	}()

	// Waiting for something not to happen is inherently hard...
	select {
	case <-acquired:
		c.Fatalf("Unexpected lock acquisition")
	case <-time.After(shortWait * 2):
		// all good.
	}
	close(kill)
	<-done
	<-remote
}

func (s *mutexSuite) TestProcessReleasesWhenDead(c *gc.C) {
	spec := s.spec()
	kill := make(chan struct{})

	// this will block until the other process has the lock.
	remote := LockFromAnotherProc(c, spec.Name, kill, false, false)

	acquired := make(chan struct{})
	done := make(chan struct{})
	go func() {
		r, err := mutex.Acquire(spec)
		if c.Check(err, gc.IsNil) {
			close(acquired)
			r.Release()
		}
		close(done)
	}()

	close(kill)

	select {
	case <-acquired:
		// all good.
	case <-time.After(appWait):
		c.Fatalf("timout waiting for mutex to be acquired")
	}
	<-done
	<-remote
}

func (s *mutexSuite) TestStress(c *gc.C) {
	const lockAttempts = 200
	const concurrentLocks = 10

	var counter = new(int64)
	// Use atomics to update lockState to make sure the lock isn't held by
	// someone else. A value of 1 means locked, 0 means unlocked.
	var lockState = new(int32)

	var wg sync.WaitGroup

	spec := s.spec()
	var stress = func() {
		defer wg.Done()
		for i := 0; i < lockAttempts; i++ {
			r, err := mutex.Acquire(spec)
			if !c.Check(err, jc.ErrorIsNil) {
				return
			}
			state := atomic.AddInt32(lockState, 1)
			if !c.Check(state, gc.Equals, int32(1)) {
				return
			}
			// Tell the go routine scheduler to give a slice to someone else
			// while we have this locked.
			runtime.Gosched()
			// need to decrement prior to unlock to avoid the race of someone
			// else grabbing the lock before we decrement the state.
			atomic.AddInt32(lockState, -1)
			r.Release()
			// increment the general counter
			atomic.AddInt64(counter, 1)
		}
	}

	for i := 0; i < concurrentLocks; i++ {
		wg.Add(1)
		go stress()
	}
	wg.Wait()
	c.Assert(*counter, gc.Equals, int64(lockAttempts*concurrentLocks))
}

// TestMutexNotInherited tests that subprocesses do not inherit mutexes.
func (s *mutexSuite) TestMutexNotInherited(c *gc.C) {
	spec := s.spec()
	r, err := mutex.Acquire(spec)
	c.Assert(err, jc.ErrorIsNil)
	defer r.Release()

	kill := make(chan struct{})
	SleepFromAnotherProc(c, kill)
	defer close(kill)
	r.Release()

	// We should be able to acquire again now.
	spec.Timeout = time.Nanosecond
	r, err = mutex.Acquire(spec)
	c.Assert(err, jc.ErrorIsNil)
	r.Release()
}

// TestFilePermissions tests that the file permissions are correct.
func (s *mutexSuite) TestFilePermissions(c *gc.C) {
	spec := s.spec()
	r, err := mutex.Acquire(spec)
	c.Assert(err, jc.ErrorIsNil)
	defer r.Release()

	filePath := filepath.Join(os.TempDir(), "juju-"+spec.Name)
	fileInfo, err := os.Stat(filePath)
	c.Assert(err, jc.ErrorIsNil)

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	c.Assert(ok, jc.IsTrue)

	current, err := user.Current()
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(fmt.Sprintf("%d", stat.Uid), gc.Equals, current.Uid)
	c.Assert(fmt.Sprintf("%d", stat.Gid), gc.Equals, current.Gid)
}

// TestFilePermissionsWithSudo tests that the file permissions are correct.
func (s *mutexSuite) TestFilePermissionsWithSudoEnvars(c *gc.C) {
	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	envion := NewMockEnvironment(ctrl)

	restore := jt.PatchValue(mutex.Envars, envion)
	defer restore()

	current, err := user.Current()
	c.Assert(err, jc.ErrorIsNil)

	exp := envion.EXPECT()
	exp.LookupEnv("SUDO_COMMAND").Return("test", true)
	exp.Getenv("SUDO_UID").Return(current.Uid)
	exp.Getenv("SUDO_GID").Return(current.Gid)

	spec := s.spec()
	r, err := mutex.Acquire(spec)
	c.Assert(err, jc.ErrorIsNil)
	defer r.Release()

	filePath := filepath.Join(os.TempDir(), "juju-"+spec.Name)
	fileInfo, err := os.Stat(filePath)
	c.Assert(err, jc.ErrorIsNil)

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	c.Assert(ok, jc.IsTrue)

	c.Assert(fmt.Sprintf("%d", stat.Uid), gc.Equals, current.Uid)
	c.Assert(fmt.Sprintf("%d", stat.Gid), gc.Equals, current.Gid)
}

// LockFromAnotherProc will launch a process and block until that process has
// created the lock file.  If we time out waiting for the other process to take
// the lock, this function will fail the current test.
func LockFromAnotherProc(c *gc.C, name string, kill chan struct{}, wait, unlock bool) (done chan struct{}) {
	listener, err := net.Listen("tcp", "localhost:0")
	c.Assert(err, jc.ErrorIsNil)
	defer listener.Close()

	cmd := exec.Command(os.Args[0], "-test.run", "TestLockFromOtherProcess")
	cmd.Env = append(
		// We must preserve os.Environ() on Windows,
		// or the subprocess will fail in weird and
		// wonderful ways.
		os.Environ(),
		"MUTEX_TEST_HELPER_WANTED=1",
		"MUTEX_TEST_HELPER_NAME="+name,
		"MUTEX_TEST_HELPER_ADDR="+listener.Addr().String(),
		"MUTEX_TEST_HELPER_WAIT="+fmt.Sprint(unlock),
		"MUTEX_TEST_HELPER_UNLOCK="+fmt.Sprint(unlock),
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
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

	haveLock := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if c.Check(err, jc.ErrorIsNil) {
			// Just close it, the signal is enough.
			conn.Close()
			close(haveLock)
		}
	}()
	select {
	case <-done:
		c.Fatalf("other process failed before acquiring lock")
	case <-haveLock:
		// all good
	case <-time.After(appWait):
		c.Fatalf("other process failed to acquire lock")
	}

	return done
}

// SleepFromAnotherProc will launch a process and cause it to sleep
// for a minute, to prove that subprocesses do not inherit the mutex.
func SleepFromAnotherProc(c *gc.C, kill chan struct{}) (done chan struct{}) {
	listener, err := net.Listen("tcp", "localhost:0")
	c.Assert(err, jc.ErrorIsNil)
	defer listener.Close()

	cmd := exec.Command(os.Args[0], "-test.run", "TestSleepFromOtherProcess")
	cmd.Env = append(
		// We must preserve os.Environ() on Windows,
		// or the subprocess will fail in weird and
		// wonderful ways.
		os.Environ(),
		"MUTEX_TEST_HELPER_WANTED=1",
		"MUTEX_TEST_HELPER_ADDR="+listener.Addr().String(),
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	c.Assert(cmd.Start(), jc.ErrorIsNil)

	done = make(chan struct{})
	go func() {
		defer close(done)
		cmd.Wait()
	}()
	go func() {
		select {
		case <-kill:
			cmd.Process.Kill()
		case <-done:
		}
	}()

	started := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if c.Check(err, jc.ErrorIsNil) {
			// Just close it, the signal is enough.
			conn.Close()
			close(started)
		}
	}()
	select {
	case <-started:
	case <-done:
		c.Fatalf("other process exited")
	case <-time.After(appWait):
		c.Fatalf("timed out waiting for other process to connect")
	}

	return done
}

func TestLockFromOtherProcess(t *testing.T) {
	if os.Getenv("MUTEX_TEST_HELPER_WANTED") == "" {
		return
	}

	name := os.Getenv("MUTEX_TEST_HELPER_NAME")
	unlock := os.Getenv("MUTEX_TEST_HELPER_UNLOCK") == "true"
	wait := os.Getenv("MUTEX_TEST_HELPER_WAIT") == "true"
	spec := mutex.Spec{
		Name:  name,
		Clock: &fakeClock{shortWait},
		Delay: shortWait,
	}

	r, err := mutex.Acquire(spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error acquiring mutex %q: %v", name, err)
		os.Exit(1)
	}
	// Connect to the first process to let it know we have the connection.
	addr := os.Getenv("MUTEX_TEST_HELPER_ADDR")
	_, err = net.Dial("tcp", addr)
	// We don't care about the connection because the server will just close the
	// connection anyway.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error notifying primary process: %v", err)
		os.Exit(1)
	}

	if wait {
		time.Sleep(longWait)
	}
	if unlock {
		r.Release()
	}
	os.Exit(0)
}

func TestSleepFromOtherProcess(t *testing.T) {
	if os.Getenv("MUTEX_TEST_HELPER_WANTED") == "" {
		return
	}
	addr := os.Getenv("MUTEX_TEST_HELPER_ADDR")
	_, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error notifying primary process: %v \n", err)
		os.Exit(1)
	}
	time.Sleep(time.Minute)
	os.Exit(0)
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
