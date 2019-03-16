/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"testing"

	"context"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

func TestTunnelManagerEventHandling(t *testing.T) {
	tcs := []struct {
		//tunnel inputs
		name   string
		repeat int
		test   func(tunnel *tunnelStub, cancel context.CancelFunc, ready, check, done chan bool) error
	}{
		{
			name:   "tunnel quits on stopped minikube",
			repeat: 1,
			test: func(tunnel *tunnelStub, cancel context.CancelFunc, ready, check, done chan bool) error {
				tunnel.mockClusterInfo = &Status{
					MinikubeState: Stopped,
				}
				glog.Info("waiting for tunnel to be ready.")
				<-ready
				glog.Info("check!")
				check <- true
				glog.Info("check done.")
				select {
				case <-done:
					glog.Info("it's done, yay!")
				case <-time.After(1 * time.Second):
					t.Error("tunnel did not stop on stopped minikube")
				}
				if tunnel.tunnelExists {
					t.Error("tunnel should not have been created")
				}
				return nil
			},
		},

		{
			name:   "tunnel quits on ctrlc before doing a check",
			repeat: 1,
			test: func(tunnel *tunnelStub, cancel context.CancelFunc, ready, check, done chan bool) error {
				tunnel.mockClusterInfo = &Status{
					MinikubeState: Stopped,
				}
				<-ready
				cancel()

				select {
				case <-done:
				case <-time.After(1 * time.Second):
					t.Error("tunnel did not stop on ctrl c")
				}

				if tunnel.tunnelExists {
					t.Error("tunnel should not have been created")
				}
				return nil
			},
		},
		{
			name:   "tunnel always quits when ctrl c is pressed",
			repeat: 100000,
			test: func(tunnel *tunnelStub, cancel context.CancelFunc, ready, check, done chan bool) error {
				tunnel.mockClusterInfo = &Status{
					MinikubeState: Running,
				}
				go func() {
					<-ready
					check <- true
					<-ready
					check <- true
					<-ready
					cancel()
					check <- true

				}()

				select {
				case <-done:
				case <-time.After(500 * time.Millisecond):
					t.Error("tunnel did not stop on ctrl c")
					return errors.New("error")
				}

				if tunnel.tunnelExists {
					t.Error("tunnel should not have been created")
					return errors.New("error")
				}

				if tunnel.timesChecked != 2 {
					t.Errorf("expected to have 2 tunnel checks, got %d", tunnel.timesChecked)
					return errors.New("error")
				}
				return nil
			},
		},
	}

	//t.Parallel()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			for i := 1; i <= tc.repeat && err == nil; i++ {
				tunnelManager := &Manager{}
				tunnel := &tunnelStub{}

				ready := make(chan bool, 1)
				check := make(chan bool, 1)
				done := make(chan bool, 1)

				ctx, cancel := context.WithCancel(context.Background())
				go tunnelManager.run(ctx, tunnel, ready, check, done)
				err = tc.test(tunnel, cancel, ready, check, done)
				if err != nil {
					glog.Errorf("error at %d", i)
				}
			}
		})

	}
}

func TestTunnelManagerDelayAndContext(t *testing.T) {
	tunnelManager := &Manager{
		delay: 1000 * time.Millisecond,
	}
	tunnel := &tunnelStub{
		mockClusterInfo: &Status{
			MinikubeState: Running,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	done, err := tunnelManager.startTunnel(ctx, tunnel)
	if err != nil {
		t.Errorf("creating tunnel failed: %s", err)
	}
	time.Sleep(1100 * time.Millisecond)
	cancel()
	<-done

	if tunnel.timesChecked < 1 {
		t.Errorf("tunnel check did not run at all")
	}

	if tunnel.timesChecked > 2 {
		t.Errorf("tunnel check ran too many times %d", tunnel.timesChecked)
	}
}

func registerRunningTunnels(reg *persistentRegistry) (*ID, *ID, error) {
	runningTunnel1 := &ID{
		Route:       unsafeParseRoute("1.2.3.4", "5.6.7.8/9"),
		Pid:         os.Getpid(),
		MachineName: "minikube",
	}

	runningTunnel2 := &ID{
		Route:       unsafeParseRoute("100.2.3.4", "200.6.7.8/9"),
		Pid:         os.Getpid(),
		MachineName: "minikube",
	}

	err := reg.Register(runningTunnel1)
	if err != nil {
		return runningTunnel1, runningTunnel2, err
	}
	err = reg.Register(runningTunnel2)
	if err != nil {
		return runningTunnel1, runningTunnel2, err
	}

	return runningTunnel1, runningTunnel2, nil
}

func registerNotRunningTunnels(reg *persistentRegistry) (*ID, *ID, error) {
	notRunningTunnel1 := &ID{
		Route:       unsafeParseRoute("200.2.3.4", "10.6.7.8/9"),
		Pid:         12341234,
		MachineName: "minikube",
	}

	notRunningTunnel2 := &ID{
		Route:       unsafeParseRoute("250.2.3.4", "20.6.7.8/9"),
		Pid:         12341234,
		MachineName: "minikube",
	}

	err := reg.Register(notRunningTunnel1)
	if err != nil {
		return notRunningTunnel1, notRunningTunnel2, err
	}
	err = reg.Register(notRunningTunnel2)
	if err != nil {
		return notRunningTunnel1, notRunningTunnel2, err
	}

	return notRunningTunnel1, notRunningTunnel2, nil
}
func TestTunnelManagerCleanup(t *testing.T) {
	reg, cleanup := createTestRegistry(t)
	defer cleanup()

	runningTunnel1, runningTunnel2, err := registerRunningTunnels(reg)
	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}

	notRunningTunnel1, notRunningTunnel2, err := registerNotRunningTunnels(reg)
	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}

	router := &fakeRouter{}

	err = router.EnsureRouteIsAdded(runningTunnel1.Route)
	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}
	err = router.EnsureRouteIsAdded(runningTunnel2.Route)
	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}
	err = router.EnsureRouteIsAdded(notRunningTunnel1.Route)
	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}
	err = router.EnsureRouteIsAdded(notRunningTunnel2.Route)
	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}

	manager := NewManager()
	manager.router = router
	manager.registry = reg

	err = manager.CleanupNotRunningTunnels()

	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}

	if len(router.rt) != 2 ||
		!router.rt[0].route.Equal(runningTunnel1.Route) ||
		!router.rt[1].route.Equal(runningTunnel2.Route) {
		t.Errorf("routes are not cleaned up, expected only running tunnels to stay, got: %s", router.rt.String())
	}

	tunnels, err := reg.List()

	if err != nil {
		t.Errorf("expected no error got: %v", err)
	}

	if len(tunnels) != 2 ||
		!tunnels[0].Equal(runningTunnel1) ||
		!tunnels[1].Equal(runningTunnel2) {
		t.Errorf("tunnels are not cleaned up properly, expected only running tunnels to stay, got: %v", tunnels)
	}

}

type tunnelStub struct {
	mockClusterInfo *Status
	tunnelExists    bool
	timesChecked    int
}

func (t *tunnelStub) update() *Status {
	t.tunnelExists = t.mockClusterInfo.MinikubeState == Running
	t.timesChecked++
	return t.mockClusterInfo
}

func (t *tunnelStub) cleanup() *Status {
	t.tunnelExists = false
	return t.mockClusterInfo
}
