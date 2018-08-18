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
	"k8s.io/minikube/pkg/minikube/constants"
	"github.com/docker/machine/libmachine/persist"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"os"
	"syscall"
	"fmt"
)

// Manager can create, start and cleanup a tunnel
// It keeps track of created tunnels for multiple vms so that it can cleanup
// after unclean shutdowns.
type Manager struct {
	delay            time.Duration
	registry         *persistentRegistry
	isProcessRunning pidChecker
	router           router
}

type pidChecker func(int) (bool, error)

func checkIfRunning(pid int) (bool, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}
	if p == nil {
		return false, nil
	}
	return true, nil
}

func NewManager() *Manager {
	return &Manager{
		delay: 5 * time.Second,
		registry: &persistentRegistry{
			fileName: constants.GetTunnelRegistryFile(),
		},
		isProcessRunning: checkIfRunning,
		router: &osRouter{},
	}
}
func (mgr *Manager) StartTunnel(ctx context.Context, machineName string,
	machineStore persist.Store,
	configLoader config.ConfigLoader,
	v1Core v1.CoreV1Interface) (done chan bool, err error) {
	tunnel, e := newTunnel(machineName, machineStore, configLoader, v1Core, mgr.registry)
	if e != nil {
		return nil, fmt.Errorf("error creating tunnel: %s", e)
	}
	return mgr.startTunnel(ctx, tunnel)

}
func (mgr *Manager) startTunnel(ctx context.Context, tunnel tunnel) (done chan bool, err error) {
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

func (mgr *Manager) run(t tunnel, ctx context.Context, ready, check, done chan bool) {
	defer func() {
		done <- true
	}()
	ready <- true
	for {
		select {
		case <-ctx.Done():
			mgr.cleanup(t)
			return
		case <-check:
			logrus.Debug("check receieved")
			select {
			case <-ctx.Done():
				mgr.cleanup(t)
				return
			default:
			}
			status := t.updateTunnelStatus()
			logrus.Debug("minikube status: %s", status)
			if status.MinikubeState != Running {
				mgr.cleanup(t)
				return
			}
			ready <- true
		}
	}
}

func (mgr *Manager) cleanup(t tunnel) *TunnelState {
	return t.cleanup()
}

func (mgr *Manager) CleanupNotRunningTunnels() error {
	logrus.Infof("cleaning up tunnels")
	tunnels, e := mgr.registry.List()
	if e != nil {
		return fmt.Errorf("error listing tunnels from registry: %s", e)
	}

	for _, tunnel := range tunnels {
		isRunning, e := mgr.isProcessRunning(tunnel.Pid)
		logrus.Infof("%v is running: %s", tunnel, isRunning)
		if e != nil {
			return e
		}
		if !isRunning {
			e = mgr.router.Cleanup(tunnel.Route)
			if e != nil {
				return e
			}
			e = mgr.registry.Remove(tunnel.Route)
			if e != nil {
				return e
			}
		}
	}
	return nil
}
