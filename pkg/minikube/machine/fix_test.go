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

package machine

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/host"
	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/libmachine/swarm"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/tests"
)

func TestFixHostProvisioning(t *testing.T) {
	for _, tc := range []struct {
		name          string
		guest         string
		state         state.State
		wantProvision bool
	}{
		{name: "stopped Windows", guest: config.GuestOSWindows, state: state.Stopped},
		{name: "running Windows", guest: config.GuestOSWindows, state: state.Running},
		{name: "stopped Linux", guest: "linux", state: state.Stopped, wantProvision: true},
		{name: "legacy Linux", state: state.Running, wantProvision: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(localpath.MinikubeHome, t.TempDir())
			oldProfile := viper.GetString(config.ProfileName)
			t.Cleanup(func() { viper.Set(config.ProfileName, oldProfile) })
			viper.Set(config.ProfileName, "restart-test")

			n := config.Node{Name: "m02", Worker: true, Guest: config.Guest{Name: tc.guest}}
			cfg := config.ClusterConfig{Name: "restart-test", Driver: driver.HyperV, Nodes: []config.Node{n}}
			d := &restartTestDriver{MockDriver: tests.MockDriver{
				T: t, CurrentState: tc.state, HostError: true, IP: "192.0.2.2",
			}}
			h := &host.Host{
				Name: config.MachineName(cfg, n), DriverName: driver.HyperV, Driver: d,
				Guest: host.Guest(n.Guest),
				HostOptions: &host.Options{
					EngineOptions: &engine.Options{},
					AuthOptions:   &auth.Options{},
					SwarmOptions:  &swarm.Options{},
				},
			}
			api := tests.NewMockAPI(t)
			api.Hosts[h.Name] = h

			got, err := fixHost(api, &cfg, &n)
			if tc.wantProvision {
				if err == nil || !strings.Contains(err.Error(), "provision: error getting host") {
					t.Fatalf("fixHost error = %v; want Linux provisioning to attempt SSH", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("fixHost attempted Linux provisioning for Windows: %v", err)
			}
			if got != h || d.CurrentState != state.Running {
				t.Fatalf("fixHost did not preserve and start the existing host: host=%p state=%v", got, d.CurrentState)
			}
			if tc.state == state.Stopped {
				if !api.SaveCalled || n.IP != d.IP {
					t.Fatalf("restarted host was not saved with its current IP: saved=%v IP=%q", api.SaveCalled, n.IP)
				}
			}
		})
	}
}

type restartTestDriver struct {
	tests.MockDriver
}

func (*restartTestDriver) DriverName() string { return driver.HyperV }
