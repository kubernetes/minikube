// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mutex

import (
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/juju/errors"
)

const prefix = "@/var/lib/juju/mutex-"

var errLocked = errors.New("already locked")

type socketMutex struct {
	mu     sync.Mutex
	socket *net.UnixListener
}

func acquireLegacy(
	name string,
	clock Clock,
	delay time.Duration,
	timeout <-chan time.Time,
	cancel <-chan struct{},
) (Releaser, error) {
	for {
		impl, err := acquireAbstractDomainSocket(name)
		if err == nil {
			return impl, nil
		} else if err != errLocked {
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

func acquireAbstractDomainSocket(name string) (Releaser, error) {
	path := filepath.Join(prefix, name)
	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return nil, errors.Trace(err)
	}
	l, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, errLocked
	}
	return &socketMutex{socket: l}, nil
}

// Release implements Releaser.
func (m *socketMutex) Release() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.socket == nil {
		return
	}
	if err := m.socket.Close(); err != nil {
		panic(err)
	}
	m.socket = nil
}
