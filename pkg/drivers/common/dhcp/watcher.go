/*
Copyright 2026 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dhcp

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	// checkInterval is how often we check as a fallback when listening for notifications.
	checkInterval = 10 * time.Second
)

type event int

const (
	fileCreated event = iota
	fileChanged
	fileMissing
	periodicCheck
	deadlineExceeded
)

// watcher monitors the DHCP leases file for changes via fsnotify, with a
// periodic fallback check every checkInterval. When the file does not exist,
// watch() returns fileMissing immediately and the caller is responsible for
// waiting before retrying.
//
// We watch the file directly rather than its parent directory (/var/db) because
// macOS SIP protects some entries under /var/db, causing fsnotify's Add() to
// fail with "operation not permitted" when it tries to lstat protected entries.
//
// Watching a file directly has one complication: programs that update the file
// atomically (write tmp + rename) cause kqueue to lose the watch on the
// original inode. We handle this by re-watching via startListening() whenever a
// Remove or Rename event is received.
//
// The periodic fallback check is an extra safety mechanism in case fsnotify
// events are buggy or silently lost.
type watcher struct {
	fsw       *fsnotify.Watcher
	deadline  *time.Timer
	ticker    *time.Ticker
	listening bool
}

func newWatcher(d time.Duration) (*watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	ticker := time.NewTicker(checkInterval)
	ticker.Stop()
	return &watcher{
		fsw:      fsw,
		deadline: time.NewTimer(d),
		ticker:   ticker,
	}, nil
}

func (w *watcher) Close() {
	w.ticker.Stop()
	w.deadline.Stop()
	w.fsw.Close()
}

// startListening tries to watch the leases file for notifications.
func (w *watcher) startListening() error {
	if err := w.fsw.Add(leasesPath); err != nil {
		if w.listening {
			w.ticker.Stop()
			w.listening = false
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to watch %s: %w", leasesPath, err)
		}
		return nil
	}
	w.listening = true
	w.ticker.Reset(checkInterval)
	return nil
}

// watch returns the next event. When the file is missing, it returns
// fileMissing immediately — the caller should sleep before calling again.
// When the file exists, it blocks until a change is detected.
func (w *watcher) watch() (event, error) {
	if !w.listening {
		select {
		case <-w.deadline.C:
			return deadlineExceeded, nil
		default:
		}
		if err := w.startListening(); err != nil {
			return 0, err
		}
		if w.listening {
			return fileCreated, nil
		}
		return fileMissing, nil
	}

	for {
		select {
		case ev, ok := <-w.fsw.Events:
			if !ok {
				return 0, errors.New("watcher events channel closed")
			}
			if ev.Has(fsnotify.Remove | fsnotify.Rename) {
				// Atomic writes replace the file via remove or rename,
				// dropping the kqueue watch. Re-watch the new inode.
				// See https://github.com/fsnotify/fsnotify/issues/372
				if err := w.startListening(); err != nil {
					return 0, err
				}
				if w.listening {
					return fileChanged, nil
				}
				return fileMissing, nil
			}
			if ev.Has(fsnotify.Create | fsnotify.Write) {
				// In practice dhcpd updates the leases file atomically
				// (write tmp + rename), so we typically see Rename above.
				// Write is kept as a fallback; if we read mid-write,
				// IPAddressForMAC simply won't find the MAC and we'll
				// retry on the next event or periodic check.
				return fileChanged, nil
			}
			// Ignore Chmod and other unrelated events (e.g. Spotlight indexing).
		case err, ok := <-w.fsw.Errors:
			if !ok {
				return 0, errors.New("watcher errors channel closed")
			}
			return 0, fmt.Errorf("watcher error: %w", err)
		case <-w.ticker.C:
			return periodicCheck, nil
		case <-w.deadline.C:
			return deadlineExceeded, nil
		}
	}
}
