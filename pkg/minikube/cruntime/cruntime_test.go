/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package cruntime

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/minikube/pkg/minikube/console"
)

func TestName(t *testing.T) {
	var tests = []struct {
		runtime string
		want    string
	}{
		{"", "Docker"},
		{"docker", "Docker"},
		{"crio", "CRI-O"},
		{"cri-o", "CRI-O"},
		{"containerd", "containerd"},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			r, err := New(Config{Type: tc.runtime})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			got := r.Name()
			if got != tc.want {
				t.Errorf("Name(%s) = %q, want: %q", tc.runtime, got, tc.want)
			}
			if !console.HasStyle(got) {
				t.Fatalf("console.HasStyle(%s): %v", got, false)
			}
		})
	}
}

func TestKubeletOptions(t *testing.T) {
	var tests = []struct {
		runtime string
		want    map[string]string
	}{
		{"docker", map[string]string{"container-runtime": "docker"}},
		{"crio", map[string]string{
			"container-runtime":          "remote",
			"container-runtime-endpoint": "/var/run/crio/crio.sock",
			"image-service-endpoint":     "/var/run/crio/crio.sock",
			"runtime-request-timeout":    "15m",
		}},
		{"containerd", map[string]string{
			"container-runtime":          "remote",
			"container-runtime-endpoint": "unix:///run/containerd/containerd.sock",
			"image-service-endpoint":     "unix:///run/containerd/containerd.sock",
			"runtime-request-timeout":    "15m",
		}},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			r, err := New(Config{Type: tc.runtime})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}

			got := r.KubeletOptions()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("KubeletOptions(%s) returned diff (-want +got):\n%s", tc.runtime, diff)
			}
		})
	}
}

type serviceState int

const (
	Exited serviceState = iota
	Running
	Restarted
)

// FakeRunner is a command runner that isn't very smart.
type FakeRunner struct {
	cmds       []string
	services   map[string]serviceState
	containers map[string]string
	t          *testing.T
}

// NewFakeRunner returns a CommandRunner which emulates a systemd host
func NewFakeRunner(t *testing.T) *FakeRunner {
	return &FakeRunner{
		services:   map[string]serviceState{},
		cmds:       []string{},
		t:          t,
		containers: map[string]string{},
	}
}

// Run a fake command!
func (f *FakeRunner) CombinedOutput(cmd string) (string, error) {
	f.cmds = append(f.cmds, cmd)

	root := false
	args := strings.Split(cmd, " ")
	bin, args := args[0], args[1:]
	f.t.Logf("bin=%s args=%v", bin, args)
	if bin == "sudo" {
		root = true
		bin, args = args[0], args[1:]
	}
	switch bin {
	case "systemctl":
		return f.systemctl(args, root)
	case "docker":
		return f.docker(args, root)
	case "crictl":
		return f.crictl(args, root)
	case "crio":
		return f.crio(args, root)
	case "containerd":
		return f.containerd(args, root)
	default:
		return "", nil
	}
}

// Run a fake command!
func (f *FakeRunner) Run(cmd string) error {
	_, err := f.CombinedOutput(cmd)
	return err
}

// docker is a fake implementation of docker
func (f *FakeRunner) docker(args []string, _ bool) (string, error) {
	switch cmd := args[0]; cmd {
	case "ps":
		// ps -a --filter="name=apiserver" --format="{{.ID}}"
		if args[1] == "-a" && strings.HasPrefix(args[2], "--filter") {
			filter := strings.Split(args[2], `"`)[1]
			fname := strings.Split(filter, "=")[1]
			ids := []string{}
			f.t.Logf("fake docker: Looking for containers matching %q", fname)
			for id, cname := range f.containers {
				if strings.Contains(cname, fname) {
					ids = append(ids, id)
				}
			}
			f.t.Logf("fake docker: Found containers: %v", ids)
			return strings.Join(ids, "\n"), nil
		}
	case "stop":
		for _, id := range args[1:] {
			f.t.Logf("fake docker: Stopping id %q", id)
			if f.containers[id] == "" {
				return "", fmt.Errorf("no such container")
			}
			delete(f.containers, id)
		}
	case "rm":
		// Skip "-f" argument
		for _, id := range args[2:] {
			f.t.Logf("fake docker: Removing id %q", id)
			if f.containers[id] == "" {
				return "", fmt.Errorf("no such container")
			}
			delete(f.containers, id)

		}
	case "version":
		if args[1] == "--format" && args[2] == "'{{.Server.Version}}'" {
			return "18.06.2-ce", nil
		}

	}
	return "", nil
}

// crio is a fake implementation of crio
func (f *FakeRunner) crio(args []string, _ bool) (string, error) {
	if args[0] == "--version" {
		return "crio version 1.13.0", nil
	}
	return "", nil
}

// containerd is a fake implementation of containerd
func (f *FakeRunner) containerd(args []string, _ bool) (string, error) {
	if args[0] == "--version" {
		return "containerd github.com/containerd/containerd v1.2.0 c4446665cb9c30056f4998ed953e6d4ff22c7c39", nil
	}
	return "", nil
}

// crictl is a fake implementation of crictl
func (f *FakeRunner) crictl(args []string, _ bool) (string, error) {
	switch cmd := args[0]; cmd {
	case "ps":
		// crictl ps -a --name=apiserver --state=Running --quiet
		if args[1] == "-a" && strings.HasPrefix(args[2], "--name") {
			fname := strings.Split(args[2], "=")[1]
			ids := []string{}
			f.t.Logf("fake crictl: Looking for containers matching %q", fname)
			for id, cname := range f.containers {
				if strings.Contains(cname, fname) {
					ids = append(ids, id)
				}
			}
			f.t.Logf("fake crictl: Found containers: %v", ids)
			return strings.Join(ids, "\n"), nil
		} else if args[1] == "-a" {
			ids := []string{}
			for id := range f.containers {
				ids = append(ids, id)
			}
			f.t.Logf("fake crictl: Found containers: %v", ids)
			return strings.Join(ids, "\n"), nil

		}
	case "stop":
		for _, id := range args[1:] {
			f.t.Logf("fake crictl: Stopping id %q", id)
			if f.containers[id] == "" {
				return "", fmt.Errorf("no such container")
			}
			delete(f.containers, id)
		}
	case "rm":
		for _, id := range args[1:] {
			f.t.Logf("fake crictl: Removing id %q", id)
			if f.containers[id] == "" {
				return "", fmt.Errorf("no such container")
			}
			delete(f.containers, id)

		}

	}
	return "", nil
}

// systemctl is a fake implementation of systemctl
func (f *FakeRunner) systemctl(args []string, root bool) (string, error) {
	action := args[0]
	svcs := args[1:]
	out := ""

	for i, arg := range args {
		// systemctl is-active --quiet service crio
		if arg == "service" {
			svcs = args[i+1:]
		}
	}

	for _, svc := range svcs {
		state, ok := f.services[svc]
		if !ok {
			return out, fmt.Errorf("unknown fake service: %s", svc)
		}

		switch action {
		case "stop":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = Exited
			f.t.Logf("fake systemctl: stopped %s", svc)
		case "start":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = Running
			f.t.Logf("fake systemctl: started %s", svc)
		case "restart":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = Restarted
			f.t.Logf("fake systemctl: restarted %s", svc)
		case "is-active":
			f.t.Logf("fake systemctl: %s is-status: %v", svc, state)
			if state == Running {
				return out, nil
			}
			return out, fmt.Errorf("%s in state: %v", svc, state)
		default:
			return out, fmt.Errorf("unimplemented fake action: %q", action)
		}
	}
	return out, nil
}

func TestVersion(t *testing.T) {
	var tests = []struct {
		runtime string
		want    string
	}{
		{"docker", "18.06.2-ce"},
		{"cri-o", "1.13.0"},
		{"containerd", "1.2.0"},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			runner := NewFakeRunner(t)
			r, err := New(Config{Type: tc.runtime, Runner: runner})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			got, err := r.Version()
			if err != nil {
				t.Fatalf("Version(%s): %v", tc.runtime, err)
			}
			if got != tc.want {
				t.Errorf("Version(%s) = %q, want: %q", tc.runtime, got, tc.want)
			}
		})
	}
}

// defaultServices reflects the default boot state for the minikube VM
var defaultServices = map[string]serviceState{
	"docker":        Running,
	"docker.socket": Running,
	"crio":          Exited,
	"crio-shutdown": Exited,
	"containerd":    Exited,
}

func TestDisable(t *testing.T) {
	var tests = []struct {
		runtime string
		want    []string
	}{
		{"docker", []string{"sudo systemctl stop docker docker.socket"}},
		{"crio", []string{"sudo systemctl stop crio"}},
		{"containerd", []string{"sudo systemctl stop containerd"}},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			runner := NewFakeRunner(t)
			for k, v := range defaultServices {
				runner.services[k] = v
			}
			cr, err := New(Config{Type: tc.runtime, Runner: runner})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			err = cr.Disable()
			if err != nil {
				t.Errorf("%s disable unexpected error: %v", tc.runtime, err)
			}
			if diff := cmp.Diff(tc.want, runner.cmds); diff != "" {
				t.Errorf("Disable(%s) commands diff (-want +got):\n%s", tc.runtime, diff)
			}
		})
	}
}

func TestEnable(t *testing.T) {
	var tests = []struct {
		runtime string
		want    map[string]serviceState
	}{
		{"docker", map[string]serviceState{
			"docker":        Running,
			"docker.socket": Running,
			"containerd":    Exited,
			"crio":          Exited,
			"crio-shutdown": Exited,
		}},
		{"containerd", map[string]serviceState{
			"docker":        Exited,
			"docker.socket": Exited,
			"containerd":    Restarted,
			"crio":          Exited,
			"crio-shutdown": Exited,
		}},
		{"crio", map[string]serviceState{
			"docker":        Exited,
			"docker.socket": Exited,
			"containerd":    Exited,
			"crio":          Restarted,
			"crio-shutdown": Exited,
		}},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			runner := NewFakeRunner(t)
			for k, v := range defaultServices {
				runner.services[k] = v
			}
			cr, err := New(Config{Type: tc.runtime, Runner: runner})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			err = cr.Enable()
			if err != nil {
				t.Errorf("%s disable unexpected error: %v", tc.runtime, err)
			}
			if diff := cmp.Diff(tc.want, runner.services); diff != "" {
				t.Errorf("service diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestContainerFunctions(t *testing.T) {
	var tests = []struct {
		runtime string
	}{
		{"docker"},
		{"crio"},
		{"containerd"},
	}

	sortSlices := cmpopts.SortSlices(func(a, b string) bool { return a < b })
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			runner := NewFakeRunner(t)
			prefix := ""
			if tc.runtime == "docker" {
				prefix = "k8s_"
			}
			runner.containers = map[string]string{
				"abc0": prefix + "apiserver",
				"fgh1": prefix + "coredns",
				"xyz2": prefix + "storage",
			}
			cr, err := New(Config{Type: tc.runtime, Runner: runner})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}

			// Get the list of apiservers
			got, err := cr.ListContainers("apiserver")
			if err != nil {
				t.Fatalf("ListContainers: %v", err)
			}
			want := []string{"abc0"}
			if !cmp.Equal(got, want) {
				t.Errorf("ListContainers(apiserver) = %v, want %v", got, want)
			}

			// Stop the containers and assert that they have disappeared
			if err := cr.StopContainers(got); err != nil {
				t.Fatalf("stop failed: %v", err)
			}
			got, err = cr.ListContainers("apiserver")
			if err != nil {
				t.Fatalf("ListContainers: %v", err)
			}
			want = nil
			if diff := cmp.Diff(got, want, sortSlices); diff != "" {
				t.Errorf("ListContainers(apiserver) unexpected results, diff (-got + want): %s", diff)
			}

			// Get the list of everything else.
			got, err = cr.ListContainers("")
			if err != nil {
				t.Fatalf("ListContainers: %v", err)
			}
			want = []string{"fgh1", "xyz2"}
			if diff := cmp.Diff(got, want, sortSlices); diff != "" {
				t.Errorf("ListContainers(apiserver) unexpected results, diff (-got + want): %s", diff)
			}

			// Kill the containers and assert that they have disappeared
			if err := cr.KillContainers(got); err != nil {
				t.Errorf("KillContainers: %v", err)
			}
			got, err = cr.ListContainers("")
			if err != nil {
				t.Fatalf("ListContainers: %v", err)
			}
			if len(got) > 0 {
				t.Errorf("ListContainers(apiserver) = %v, want 0 items", got)
			}
		})
	}
}
