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
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
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
		})
	}
}

func TestImageExists(t *testing.T) {
	var tests = []struct {
		runtime string
		name    string
		sha     string
		want    bool
	}{
		{"docker", "missing-image", "0000000000000000000000000000000000000000000000000000000000000000", false},
		{"docker", "available-image", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", true},
		{"crio", "missing-image", "0000000000000000000000000000000000000000000000000000000000000000", false},
		{"crio", "available-image", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", true},
	}
	for _, tc := range tests {
		runner := NewFakeRunner(t)
		runner.images = map[string]string{
			"available-image": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		}
		t.Run(tc.runtime, func(t *testing.T) {

			r, err := New(Config{Type: tc.runtime, Runner: runner})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}

			got := r.ImageExists(tc.name, tc.sha)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ImageExists(%s) returned diff (-want +got):\n%s", tc.runtime, diff)
			}
		})
	}
}

func TestCGroupDriver(t *testing.T) {
	var tests = []struct {
		runtime string
		want    string
	}{
		{"docker", "cgroupfs"},
		{"crio", "cgroupfs"},
		{"containerd", "cgroupfs"},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			r, err := New(Config{Type: tc.runtime, Runner: NewFakeRunner(t)})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}

			got, err := r.CGroupDriver()
			if err != nil {
				t.Fatalf("CGroupDriver(): %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("CGroupDriver(%s) returned diff (-want +got):\n%s", tc.runtime, diff)
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
	SvcExited serviceState = iota
	SvcRunning
	SvcRestarted
)

// FakeRunner is a command runner that isn't very smart.
type FakeRunner struct {
	cmds       []string
	services   map[string]serviceState
	containers map[string]string
	images     map[string]string
	t          *testing.T
}

// NewFakeRunner returns a CommandRunner which emulates a systemd host
func NewFakeRunner(t *testing.T) *FakeRunner {
	return &FakeRunner{
		services:   map[string]serviceState{},
		cmds:       []string{},
		t:          t,
		containers: map[string]string{},
		images:     map[string]string{},
	}
}

func buffer(s string, err error) (*command.RunResult, error) {
	rr := &command.RunResult{}
	if err != nil {
		return rr, err
	}
	var buf bytes.Buffer
	_, err = buf.WriteString(s)
	if err != nil {
		return rr, errors.Wrap(err, "Writing outStr to FakeRunner's buffer")
	}
	rr.Stdout = buf
	rr.Stderr = buf
	return rr, err
}

// Run a fake command!
func (f *FakeRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	xargs := cmd.Args
	f.cmds = append(f.cmds, xargs...)
	root := false
	bin, args := xargs[0], xargs[1:]
	f.t.Logf("bin=%s args=%v", bin, args)
	if bin == "sudo" {
		root = true
		bin, args = xargs[1], xargs[2:]
	}
	switch bin {
	case "systemctl":
		return buffer(f.systemctl(args, root))
	case "which":
		return buffer(f.which(args, root))
	case "docker":
		return buffer(f.docker(args, root))
	case "podman":
		return buffer(f.podman(args, root))
	case "crictl", "/usr/bin/crictl":
		return buffer(f.crictl(args, root))
	case "crio":
		return buffer(f.crio(args, root))
	case "containerd":
		return buffer(f.containerd(args, root))
	default:
		rr := &command.RunResult{}
		return rr, nil
	}
}

func (f *FakeRunner) StartCmd(cmd *exec.Cmd) (*command.StartedCmd, error) {
	return &command.StartedCmd{}, nil
}

func (f *FakeRunner) WaitCmd(sc *command.StartedCmd) (*command.RunResult, error) {
	return &command.RunResult{}, nil
}

func (f *FakeRunner) Copy(assets.CopyableFile) error {
	return nil
}

func (f *FakeRunner) Remove(assets.CopyableFile) error {
	return nil
}

func (f *FakeRunner) dockerPs(args []string) (string, error) {
	// ps -a --filter="name=apiserver" --format="{{.ID}}"
	if args[1] == "-a" && strings.HasPrefix(args[2], "--filter") {
		filter := strings.Split(args[2], `r=`)[1]
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
	return "", nil
}

func (f *FakeRunner) dockerStop(args []string) (string, error) {
	ids := strings.Split(args[1], " ")
	for _, id := range ids {
		f.t.Logf("fake docker: Stopping id %q", id)
		if f.containers[id] == "" {
			return "", fmt.Errorf("no such container")
		}
		delete(f.containers, id)
	}
	return "", nil
}

func (f *FakeRunner) dockerRm(args []string) (string, error) {
	// Skip "-f" argument
	for _, id := range args[2:] {
		f.t.Logf("fake docker: Removing id %q", id)
		if f.containers[id] == "" {
			return "", fmt.Errorf("no such container")
		}
		delete(f.containers, id)
	}
	return "", nil
}

func (f *FakeRunner) dockerInspect(args []string) (string, error) {
	if args[1] == "--format" && args[2] == "{{.Id}}" {
		image, ok := f.images[args[3]]
		if !ok {
			return "", &exec.ExitError{Stderr: []byte("Error: No such object: missing")}
		}
		return "sha256:" + image, nil
	}
	return "", nil
}

func (f *FakeRunner) dockerRmi(args []string) (string, error) {
	// Skip "-f" argument
	for _, id := range args[1:] {
		f.t.Logf("fake docker: Removing id %q", id)
		if f.images[id] == "" {
			return "", fmt.Errorf("no such image")
		}
		delete(f.images, id)
	}
	return "", nil
}

// docker is a fake implementation of docker
func (f *FakeRunner) docker(args []string, _ bool) (string, error) {
	switch cmd := args[0]; cmd {
	case "ps":
		return f.dockerPs(args)

	case "stop":
		return f.dockerStop(args)

	case "rm":
		return f.dockerRm(args)

	case "version":

		if args[1] == "--format" && args[2] == "{{.Server.Version}}" {
			return "18.06.2-ce", nil
		}

	case "image":
		if args[1] == "inspect" {
			return f.dockerInspect(args[1:])
		}

	case "rmi":
		return f.dockerRmi(args)

	case "inspect":
		return f.dockerInspect(args)

	case "info":

		if args[1] == "--format" && args[2] == "{{.CgroupDriver}}" {
			return "cgroupfs", nil
		}
	}
	return "", nil
}

// podman is a fake implementation of podman
func (f *FakeRunner) podman(args []string, _ bool) (string, error) {
	switch cmd := args[0]; cmd {
	case "--version":
		return "podman version 1.6.4", nil

	case "image":

		if args[1] == "inspect" && args[2] == "--format" && args[3] == "{{.Id}}" {
			if args[3] == "missing" {
				return "", &exec.ExitError{Stderr: []byte("Error: error getting image \"missing\": unable to find a name and tag match for missing in repotags: no such image")}
			}
			return "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", nil
		}

	}
	return "", nil
}

// crio is a fake implementation of crio
func (f *FakeRunner) crio(args []string, _ bool) (string, error) { //nolint (result 1 (error) is always nil)
	if args[0] == "--version" {
		return "crio version 1.13.0", nil
	}
	if args[0] == "config" {
		return "# Cgroup management implementation used for the runtime.\ncgroup_manager = \"cgroupfs\"\n", nil
	}
	return "", nil
}

// containerd is a fake implementation of containerd
func (f *FakeRunner) containerd(args []string, _ bool) (string, error) {
	if args[0] == "--version" {
		return "containerd github.com/containerd/containerd v1.2.0 c4446665cb9c30056f4998ed953e6d4ff22c7c39", nil
	}
	if args[0] != "--version" { // doing this to suppress lint "result 1 (error) is always nil"
		return "", fmt.Errorf("unknown args[0]")
	}
	return "", nil
}

// crictl is a fake implementation of crictl
func (f *FakeRunner) crictl(args []string, _ bool) (string, error) {
	f.t.Logf("crictl args: %s", args)
	switch cmd := args[0]; cmd {
	case "info":
		return `{
		  "status": {
		  },
		  "config": {
		    "systemdCgroup": false
		  },
		  "golang": "go1.11.13"
		}`, nil
	case "ps":
		fmt.Printf("args %d: %v\n", len(args), args)
		if len(args) != 4 {
			f.t.Logf("crictl all")
			ids := []string{}
			for id := range f.containers {
				ids = append(ids, id)
			}
			f.t.Logf("fake crictl: Found containers: %v", ids)
			return strings.Join(ids, "\n"), nil
		}

		// crictl ps -a --name=apiserver --state=Running --quiet
		if args[1] == "-a" && strings.HasPrefix(args[3], "--name") {
			fname := strings.Split(args[3], "=")[1]
			f.t.Logf("crictl filter for %s", fname)
			ids := []string{}
			f.t.Logf("fake crictl: Looking for containers matching %q", fname)
			for id, cname := range f.containers {
				if strings.Contains(cname, fname) {
					ids = append(ids, id)
				}
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
	case "rmi":
		for _, id := range args[1:] {
			f.t.Logf("fake crictl: Removing id %q", id)
			if f.images[id] == "" {
				return "", fmt.Errorf("no such image")
			}
			delete(f.images, id)
		}
	}
	return "", nil
}

// systemctl is a fake implementation of systemctl
func (f *FakeRunner) systemctl(args []string, root bool) (string, error) { // nolint result 0 (string) is always ""
	klog.Infof("fake systemctl: %v", args)
	action := args[0]

	if action == "--version" {
		return "systemd 123 (321.2-1)", nil
	}

	if action == "daemon-reload" {
		return "ok", nil
	}

	var svcs []string
	if len(args) > 0 {
		svcs = args[1:]
	}

	// force
	if svcs[0] == "-f" {
		svcs = svcs[1:]
	}

	out := ""

	for i, arg := range args {
		// systemctl is-active --quiet service crio
		if arg == "service" {
			svcs = args[i+1:]
		}

	}

	for _, svc := range svcs {
		svc = strings.Replace(svc, ".service", "", 1)
		state, ok := f.services[svc]
		if !ok {
			return out, fmt.Errorf("unknown fake service: %s", svc)
		}

		switch action {
		case "stop":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = SvcExited
			f.t.Logf("fake systemctl: stopped %s", svc)
		case "start":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = SvcRunning
			f.t.Logf("fake systemctl: started %s", svc)
		case "restart":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = SvcRestarted
			f.t.Logf("fake systemctl: SvcRestarted %s", svc)
		case "is-active":
			f.t.Logf("fake systemctl: %s is-status: %v", svc, state)
			if state == SvcRunning {
				return out, nil
			}
			return out, fmt.Errorf("%s in state: %v", svc, state)
		case "cat":
			f.t.Logf("fake systemctl: %s cat: %v", svc, state)
			if svc == "docker.service" {
				out += "[Unit]\n"
				out += "Description=Docker Application Container Engine\n"
				out += "Documentation=https://docs.docker.com\n"
				// out += "BindsTo=containerd.service\n"
				return out, nil
			}
			return out, fmt.Errorf("%s cat unimplemented", svc)
		case "enable":
		case "disable":
		case "mask":
		case "unmask":
			f.t.Logf("fake systemctl: %s %s: %v", svc, action, state)
		default:
			return out, fmt.Errorf("unimplemented fake action: %q", action)
		}
	}
	return out, nil
}

// which is a fake implementation of which
func (f *FakeRunner) which(args []string, root bool) (string, error) { // nolint result 0 (string) is always ""
	command := args[0]
	path := fmt.Sprintf("/usr/bin/%s", command)
	return path, nil
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
			r, err := New(Config{Type: tc.runtime, Runner: NewFakeRunner(t)})
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
	"docker":        SvcRunning,
	"crio":          SvcExited,
	"crio-shutdown": SvcExited,
	"containerd":    SvcExited,
}

// allServices reflects the state of all actual services running at once
var allServices = map[string]serviceState{
	"docker":        SvcRunning,
	"crio":          SvcRunning,
	"crio-shutdown": SvcExited,
	"containerd":    SvcRunning,
}

func TestDisable(t *testing.T) {
	var tests = []struct {
		runtime string
		want    []string
	}{
		{"docker", []string{"sudo", "systemctl", "stop", "-f", "docker.socket", "sudo", "systemctl", "stop", "-f", "docker.service",
			"sudo", "systemctl", "disable", "docker.socket", "sudo", "systemctl", "mask", "docker.service"}},
		{"crio", []string{"sudo", "systemctl", "stop", "-f", "crio"}},
		{"containerd", []string{"sudo", "systemctl", "stop", "-f", "containerd"}},
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
		runtime  string
		services map[string]serviceState
		want     map[string]serviceState
	}{
		{"docker", defaultServices,
			map[string]serviceState{
				"docker":        SvcRunning,
				"containerd":    SvcExited,
				"crio":          SvcExited,
				"crio-shutdown": SvcExited,
			}},
		{"docker", allServices,
			map[string]serviceState{
				"docker":        SvcRestarted,
				"containerd":    SvcExited,
				"crio":          SvcExited,
				"crio-shutdown": SvcExited,
			}},
		{"containerd", defaultServices,
			map[string]serviceState{
				"docker":        SvcExited,
				"containerd":    SvcRestarted,
				"crio":          SvcExited,
				"crio-shutdown": SvcExited,
			}},
		{"crio", defaultServices,
			map[string]serviceState{
				"docker":        SvcExited,
				"containerd":    SvcExited,
				"crio":          SvcRunning,
				"crio-shutdown": SvcExited,
			}},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			runner := NewFakeRunner(t)
			for k, v := range tc.services {
				runner.services[k] = v
			}
			cr, err := New(Config{Type: tc.runtime, Runner: runner})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			err = cr.Enable(true, false)
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
			runner.images = map[string]string{
				"image1": "latest",
			}
			cr, err := New(Config{Type: tc.runtime, Runner: runner})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}

			// Get the list of apiservers
			got, err := cr.ListContainers(ListContainersOptions{Name: "apiserver"})
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
			got, err = cr.ListContainers(ListContainersOptions{Name: "apiserver"})
			if err != nil {
				t.Fatalf("ListContainers: %v", err)
			}
			want = nil
			if diff := cmp.Diff(got, want, sortSlices); diff != "" {
				t.Errorf("ListContainers(apiserver) unexpected results, diff (-got + want): %s", diff)
			}

			// Get the list of everything else.
			got, err = cr.ListContainers(ListContainersOptions{})
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
			got, err = cr.ListContainers(ListContainersOptions{})
			if err != nil {
				t.Fatalf("ListContainers: %v", err)
			}
			if len(got) > 0 {
				t.Errorf("ListContainers(apiserver) = %v, want 0 items", got)
			}

			// Remove a image
			if err := cr.RemoveImage("image1"); err != nil {
				t.Fatalf("RemoveImage: %v", err)
			}
			if len(runner.images) > 0 {
				t.Errorf("RemoveImage = %v, want 0 items", len(runner.images))
			}
		})
	}
}
