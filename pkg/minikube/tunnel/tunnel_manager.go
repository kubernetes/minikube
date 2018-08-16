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
	"time"

	"context"
	"github.com/sirupsen/logrus"
	"k8s.io/minikube/pkg/minikube/tunnel/types"
)

// Manager can create, start and cleanup a tunnel
// It keeps track of created tunnels for multiple vms so that it can cleanup
// after unclean shutdowns.
type Manager struct {
	delay time.Duration
	registry *registry
}

func NewManager() *Manager {
	return &Manager{
		delay: 5 * time.Second,
	}
}

func (mgr *Manager) StartTunnel(ctx context.Context, tunnel Tunnel) (done chan bool, err error) {
	logrus.Infof("Setting up tunnel...")

	ready := make(chan bool, 1)
	check := make(chan bool, 1)
	done = make(chan bool, 1)

	//simulating Ctrl+C so that we can cancel the tunnel programmatically too
	go mgr.timerLoop(ready, check)
	go mgr.run(tunnel, ctx, ready, check, done)

	logrus.Infof("Started minikube tunnel.")

	return
}

func (mgr *Manager) timerLoop(ready, check chan bool) {
	for {
		logrus.Debugf("waiting for tunnel to be ready for next check")
		<-ready
		logrus.Debugf("sleep for %s", mgr.delay)
		time.Sleep(mgr.delay)
		check <- true
	}
}

func (mgr *Manager) run(t Tunnel, ctx context.Context, ready, check, done chan bool) {
	defer func() {
		done <- true
	}()
	ready <- true
	for {
		select {
		case <-ctx.Done():
			t.cleanup()
			return
		case <-check:
			logrus.Debug("check receieved")
			select {
			case <-ctx.Done():
				t.cleanup()
				return
			default:
			}
			status := t.updateTunnelStatus()
			logrus.Debug("minikube status: %s", status)
			if status.MinikubeState != types.Running {
				t.cleanup()
				return
			}
			ready <- true
		}
	}
}
