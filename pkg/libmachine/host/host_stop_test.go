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

package host

import (
	"testing"
	"time"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/mcnerror"
	"k8s.io/minikube/pkg/libmachine/state"
)

// stopFakeDriver simulates initiate-and-return Stop semantics: Stop never
// changes the state by itself, the test controls it via cooperative.
type stopFakeDriver struct {
	drivers.Driver
	state       state.State
	cooperative bool
	stopCalls   int
	killCalls   int
}

func (f *stopFakeDriver) GetState() (state.State, error) {
	return f.state, nil
}

func (d *stopFakeDriver) Stop() error {
	d.stopCalls++
	if d.cooperative {
		d.state = state.Stopped
	}
	return nil
}

func (d *stopFakeDriver) Kill() error {
	d.killCalls++
	d.state = state.Stopped
	return nil
}

func shrinkStopWait(t *testing.T) {
	t.Helper()
	oldAttempts, oldInterval := stopWaitAttempts, stopWaitInterval
	stopWaitAttempts, stopWaitInterval = 3, time.Millisecond
	t.Cleanup(func() {
		stopWaitAttempts, stopWaitInterval = oldAttempts, oldInterval
	})
}

func TestHostStopGraceful(t *testing.T) {
	shrinkStopWait(t)
	d := &stopFakeDriver{state: state.Running, cooperative: true}
	h := &Host{Name: "test", Driver: d}

	if err := h.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if d.stopCalls != 1 {
		t.Fatalf("expected 1 Stop call, got %d", d.stopCalls)
	}
	if d.killCalls != 0 {
		t.Fatalf("Kill must not be called on graceful stop, got %d calls", d.killCalls)
	}
}

func TestHostStopKillsStuckMachine(t *testing.T) {
	shrinkStopWait(t)
	d := &stopFakeDriver{state: state.Running}
	h := &Host{Name: "test", Driver: d}

	if err := h.Stop(); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if d.killCalls != 1 {
		t.Fatalf("expected Kill to be called once after the wait timed out, got %d", d.killCalls)
	}
}

func TestHostStopAlreadyStopped(t *testing.T) {
	shrinkStopWait(t)
	d := &stopFakeDriver{state: state.Stopped}
	h := &Host{Name: "test", Driver: d}

	err := h.Stop()
	if _, ok := err.(mcnerror.ErrHostAlreadyInState); !ok {
		t.Fatalf("expected ErrHostAlreadyInState, got %v", err)
	}
	if d.stopCalls+d.killCalls != 0 {
		t.Fatalf("neither Stop nor Kill may run on a stopped host")
	}
}
