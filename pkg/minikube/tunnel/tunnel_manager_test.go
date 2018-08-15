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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/minikube/tunnel/types"
	"time"
)

func TestTunnelManagerEventHandling(t *testing.T) {
	tcs := []struct {
		//tunnel inputs
		name   string
		repeat int
		store  *tests.FakeStore
		test   func(tunnel *tunnelStub, cancel context.CancelFunc, ready, check, done chan bool) error
	}{
		{
			name:   "tunnel quits on stopped minikube",
			repeat: 1,
			test: func(tunnel *tunnelStub, cancel context.CancelFunc, ready, check, done chan bool) error {
				tunnel.mockClusterInfo = &types.TunnelState{
					MinikubeState: types.Stopped,
				}
				logrus.Info("waiting for tunnel to be ready.")
				<-ready
				logrus.Info("check!")
				check <- true
				logrus.Info("check done.")
				select {
				case <-done:
					logrus.Info("it's done, yay!")
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
				tunnel.mockClusterInfo = &types.TunnelState{
					MinikubeState: types.Stopped,
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
				tunnel.mockClusterInfo = &types.TunnelState{
					MinikubeState: types.Running,
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
				go tunnelManager.run(tunnel, ctx, ready, check, done)
				err = tc.test(tunnel, cancel, ready, check, done)
				if err != nil {
					logrus.Errorf("error at %d", i)
				}
			}
		})

	}
}

func TestTunnelManagerDelayAndContext(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	tunnelManager := &Manager{
		delay: 1000 * time.Millisecond,
	}
	tunnel := &tunnelStub{
		mockClusterInfo: &types.TunnelState{
			MinikubeState: types.Running,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	done, err := tunnelManager.StartTunnel(ctx, tunnel)
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

func TestTunnelManagerCleanup(t *testing.T) {
	//inject fake registry and fake Pid inspector
	//expect
	// 	call router.cleanup on all the tunnels that have a non-running Pid
	//	print warning on all the routes that have a running Pid

}

type tunnelStub struct {
	mockClusterInfo *types.TunnelState
	mockError       error
	tunnelExists    bool
	timesChecked    int
}

func (t *tunnelStub) updateTunnelStatus() *types.TunnelState {
	t.tunnelExists = t.mockClusterInfo.MinikubeState == types.Running
	t.timesChecked += 1
	return t.mockClusterInfo
}

func (t *tunnelStub) cleanup() *types.TunnelState {
	t.tunnelExists = false
	return t.mockClusterInfo
}
