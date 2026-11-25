//go:build linux

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
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"

	_ "k8s.io/minikube/pkg/minikube/registry/drvs/none"
)

func TestLoadNoneUsesProfileRuntimeWhenMachineConfigOmitsIt(t *testing.T) {
	home := t.TempDir()
	const name = "minikube"
	writeProfile(t, home, name, driver.None, "containerd", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
		Worker:            true,
	}})
	// Clusters created before none learned UnmarshalJSON have no ContainerRuntime.
	writeMachineConfig(t, home, name, driver.None, `{"MachineName":"minikube","URL":"tcp://127.0.0.1:2376"}`)

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

func TestLoadNoneKeepsRecordedRuntime(t *testing.T) {
	home := t.TempDir()
	const name = "minikube"
	writeProfile(t, home, name, driver.None, "crio", []config.Node{{
		KubernetesVersion: "v1.31.0",
		ControlPlane:      true,
	}})
	writeMachineConfig(t, home, name, driver.None, `{"MachineName":"minikube","ContainerRuntime":"containerd"}`)

	api := newTestClient(t, home)
	h, err := api.Load(name)
	if err != nil {
		t.Fatal(err)
	}
	if got := runtimeName(t, h.Driver); got != "containerd" {
		t.Errorf("runtime = %q, want containerd", got)
	}
}
