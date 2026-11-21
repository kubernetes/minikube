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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/run"

	_ "k8s.io/minikube/pkg/minikube/registry/drvs/ssh"
)

func TestContainerRuntimeFromRaw(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		want     string
		recorded bool
	}{
		{name: "absent", raw: `{"MachineName":"issue01"}`, recorded: false},
		{name: "empty", raw: `{"ContainerRuntime":""}`, want: "", recorded: true},
		{name: "containerd", raw: `{"ContainerRuntime":"containerd"}`, want: "containerd", recorded: true},
		{name: "case", raw: `{"containerruntime":"crio"}`, want: "crio", recorded: true},
		{name: "invalid json", raw: `{`, recorded: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, recorded := containerRuntimeFromRaw([]byte(tc.raw))
			if recorded != tc.recorded || got != tc.want {
				t.Errorf("containerRuntimeFromRaw() = (%q, %v), want (%q, %v)", got, recorded, tc.want, tc.recorded)
			}
		})
	}
}

type nameOnlyDriver struct {
	name string
}

func (n nameOnlyDriver) DriverName() string { return n.name }

type stubRuntimeDriver struct {
	name    string
	applied string
	err     error
	calls   int
}

func (s *stubRuntimeDriver) DriverName() string { return s.name }
func (s *stubRuntimeDriver) UseContainerRuntime(name string) error {
	s.calls++
	s.applied = name
	return s.err
}

func TestProfileContainerRuntime(t *testing.T) {
	home := t.TempDir()
	writeProfile(t, home, "issue01", driver.SSH, "containerd", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
	}})
	writeProfile(t, home, "clust", driver.None, "crio", []config.Node{
		{Name: "", KubernetesVersion: "v1.31.0", ControlPlane: true},
		{Name: "m02", KubernetesVersion: "v1.31.0", Worker: true},
	})
	writeProfile(t, home, "olddocker", driver.SSH, "", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
	}})
	writeProfile(t, home, "generic", driver.AliasSSH, "containerd", nil)
	writeProfile(t, home, "dock", driver.Docker, "containerd", []config.Node{
		{Name: "", KubernetesVersion: "v1.31.0", ControlPlane: true},
		{Name: "m02", KubernetesVersion: "v1.31.0", Worker: true},
	})

	tests := []struct {
		name          string
		machine       string
		machineDriver string
		want          string
		ok            bool
	}{
		{name: "single node ssh", machine: "issue01", machineDriver: driver.SSH, want: "containerd", ok: true},
		{name: "multi node", machine: "clust-m02", machineDriver: driver.None, want: "crio", ok: true},
		{name: "primary of multi node", machine: "clust", machineDriver: driver.None, want: "crio", ok: true},
		{name: "empty runtime is docker", machine: "olddocker", machineDriver: driver.SSH, ok: false},
		{name: "generic alias", machine: "generic", machineDriver: driver.SSH, want: "containerd", ok: true},
		{name: "other driver profile", machine: "dock-m02", machineDriver: driver.SSH, ok: false},
		{name: "missing", machine: "nope", machineDriver: driver.SSH, ok: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := profileContainerRuntime(home, tc.machine, tc.machineDriver)
			if ok != tc.ok || got != tc.want {
				t.Errorf("profileContainerRuntime() = (%q, %v), want (%q, %v)", got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestApplyProfileContainerRuntime(t *testing.T) {
	home := t.TempDir()
	writeProfile(t, home, "issue01", driver.SSH, "containerd", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
	}})

	t.Run("legacy config uses profile", func(t *testing.T) {
		stub := &stubRuntimeDriver{name: driver.SSH}
		raw := []byte(`{"MachineName":"issue01"}`)
		if err := applyProfileContainerRuntime(stub, raw, "issue01", home); err != nil {
			t.Fatal(err)
		}
		if stub.calls != 1 || stub.applied != "containerd" {
			t.Fatalf("applied %q %d times, want containerd once", stub.applied, stub.calls)
		}
	})

	t.Run("recorded runtime is left alone", func(t *testing.T) {
		stub := &stubRuntimeDriver{name: driver.SSH}
		raw := []byte(`{"MachineName":"issue01","ContainerRuntime":"docker"}`)
		if err := applyProfileContainerRuntime(stub, raw, "issue01", home); err != nil {
			t.Fatal(err)
		}
		if stub.calls != 0 {
			t.Fatalf("UseContainerRuntime called %d times, want 0", stub.calls)
		}
	})

	t.Run("explicit empty runtime is left alone", func(t *testing.T) {
		stub := &stubRuntimeDriver{name: driver.SSH}
		raw := []byte(`{"MachineName":"issue01","ContainerRuntime":""}`)
		if err := applyProfileContainerRuntime(stub, raw, "issue01", home); err != nil {
			t.Fatal(err)
		}
		if stub.calls != 0 {
			t.Fatalf("UseContainerRuntime called %d times, want 0", stub.calls)
		}
	})

	t.Run("driver without a runtime manager", func(t *testing.T) {
		// Drivers other than ssh and none do not implement containerRuntimeUser.
		plain := nameOnlyDriver{name: driver.SSH}
		if err := applyProfileContainerRuntime(plain, []byte(`{"MachineName":"issue01"}`), "issue01", home); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid profile runtime", func(t *testing.T) {
		badHome := t.TempDir()
		writeProfile(t, badHome, "issue01", driver.SSH, "not-a-runtime", []config.Node{{
			KubernetesVersion: "v1.31.0",
			ControlPlane:      true,
		}})
		stub := &stubRuntimeDriver{name: driver.SSH, err: fmt.Errorf("unknown runtime type")}
		err := applyProfileContainerRuntime(stub, []byte(`{"MachineName":"issue01"}`), "issue01", badHome)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestLoadSSHUsesProfileRuntimeWhenMachineConfigOmitsIt(t *testing.T) {
	home := t.TempDir()
	const name = "issue01"
	writeProfile(t, home, name, driver.SSH, "containerd", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
		Worker:            true,
	}})
	// Shape of a machine config written before ContainerRuntime existed.
	writeMachineConfig(t, home, name, driver.SSH, `{"MachineName":"issue01","IPAddress":"192.0.2.10","SSHUser":"user","SSHPort":22}`)

	api := newTestClient(t, home)
	h, err := api.Load(name)
	if err != nil {
		t.Fatal(err)
	}
	if got := runtimeName(t, h.Driver); got != "containerd" {
		t.Errorf("runtime = %q, want containerd", got)
	}
	buf, err := json.Marshal(h.Driver)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buf), `"ContainerRuntime":"containerd"`) {
		t.Errorf("loaded driver = %s, want ContainerRuntime recorded for the next save", buf)
	}
}

func TestNewHostSSHKeepsRecordedRuntime(t *testing.T) {
	home := t.TempDir()
	const name = "issue01"
	writeProfile(t, home, name, driver.SSH, "containerd", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
	}})

	api := newTestClient(t, home)
	raw := []byte(`{"MachineName":"issue01","ContainerRuntime":"docker","IPAddress":"192.0.2.10","SSHUser":"user","SSHPort":22}`)
	h, err := api.NewHost(driver.SSH, raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := runtimeName(t, h.Driver); got != "Docker" {
		t.Errorf("runtime = %q, want Docker", got)
	}
}

func TestLoadSSHKeepsDockerWhenProfileRuntimeIsEmpty(t *testing.T) {
	home := t.TempDir()
	const name = "olddocker"
	writeProfile(t, home, name, driver.SSH, "", []config.Node{{
		KubernetesVersion: "v1.25.0",
		ControlPlane:      true,
	}})
	writeMachineConfig(t, home, name, driver.SSH, `{"MachineName":"olddocker","IPAddress":"192.0.2.10","SSHUser":"user","SSHPort":22}`)

	api := newTestClient(t, home)
	h, err := api.Load(name)
	if err != nil {
		t.Fatal(err)
	}
	// "" on a pre-2022 profile means Docker. Do not switch those clusters.
	if got := runtimeName(t, h.Driver); got != "Docker" {
		t.Errorf("runtime = %q, want Docker", got)
	}
}

func TestLoadSSHKeepsRecordedRuntime(t *testing.T) {
	home := t.TempDir()
	const name = "issue01"
	writeProfile(t, home, name, driver.SSH, "containerd", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
	}})
	writeMachineConfig(t, home, name, driver.SSH, `{"MachineName":"issue01","ContainerRuntime":"docker","IPAddress":"192.0.2.10","SSHUser":"user","SSHPort":22}`)

	api := newTestClient(t, home)
	h, err := api.Load(name)
	if err != nil {
		t.Fatal(err)
	}
	if got := runtimeName(t, h.Driver); got != "Docker" {
		t.Errorf("runtime = %q, want Docker", got)
	}
}

func TestNewHostSSHUsesProfileRuntimeWhenMachineConfigOmitsIt(t *testing.T) {
	home := t.TempDir()
	const name = "issue01"
	writeProfile(t, home, name, driver.SSH, "crio", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
	}})

	api := newTestClient(t, home)
	raw := []byte(`{"MachineName":"issue01","IPAddress":"192.0.2.10","SSHUser":"user","SSHPort":22}`)
	h, err := api.NewHost(driver.SSH, raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := runtimeName(t, h.Driver); got != "CRI-O" {
		t.Errorf("runtime = %q, want CRI-O", got)
	}
}

func TestLoadSSHInvalidProfileRuntime(t *testing.T) {
	home := t.TempDir()
	const name = "issue01"
	writeProfile(t, home, name, driver.SSH, "not-a-runtime", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
	}})
	writeMachineConfig(t, home, name, driver.SSH, `{"MachineName":"issue01","IPAddress":"192.0.2.10","SSHUser":"user","SSHPort":22}`)

	api := newTestClient(t, home)
	if _, err := api.Load(name); err == nil {
		t.Fatal("expected error for invalid profile runtime")
	}
}

func newTestClient(t *testing.T, home string) *LocalClient {
	t.Helper()
	api, err := NewAPIClient(&run.CommandOptions{}, home)
	if err != nil {
		t.Fatal(err)
	}
	client, ok := api.(*LocalClient)
	if !ok {
		t.Fatalf("API is %T, want *LocalClient", api)
	}
	return client
}

func runtimeName(t *testing.T, drv interface{}) string {
	t.Helper()
	namer, ok := drv.(interface{ RuntimeName() string })
	if !ok {
		t.Fatalf("driver %T does not report a runtime", drv)
	}
	return namer.RuntimeName()
}

func writeProfile(t *testing.T, home, name, drv, runtime string, nodes []config.Node) {
	t.Helper()
	cc := &config.ClusterConfig{
		Name:   name,
		Driver: drv,
		KubernetesConfig: config.KubernetesConfig{
			ContainerRuntime: runtime,
		},
		Nodes: nodes,
	}
	if err := config.SaveProfile(name, cc, home); err != nil {
		t.Fatal(err)
	}
}

func writeMachineConfig(t *testing.T, home, name, driverName, driverBody string) {
	t.Helper()
	dir := filepath.Join(home, "machines", name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := fmt.Sprintf(`{
  "ConfigVersion": 3,
  "Name": %q,
  "DriverName": %q,
  "Driver": %s,
  "HostOptions": {
    "AuthOptions": {
      "StorePath": %q
    }
  }
}`, name, driverName, driverBody, dir)
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(cfg), 0600); err != nil {
		t.Fatal(err)
	}
}
