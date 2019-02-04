package cruntime

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestName(t *testing.T) {
	var tests = []struct {
		runtime string
		want    string
	}{
		{"", "Docker"},
		{"docker", "Docker"},
		{"crio", "CRIO"},
		{"cri-o", "CRIO"},
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
)

// fakeHost is a command runner that isn't very smart.
type fakeHost struct {
	cmds     []string
	services map[string]serviceState
}

// Run a fake command!
func (f *fakeHost) CombinedOutput(cmd string) (string, error) {
	f.cmds = append(f.cmds, cmd)
	out := ""

	root := false
	args := strings.Split(cmd, " ")
	bin, args := args[0], args[1:]
	if bin == "sudo" {
		root = true
		bin, args = args[0], args[1:]
	}
	if bin == "systemctl" {
		return f.systemctl(args, root)
	}
	if bin == "docker" {
		return f.docker(args, root)
	}
	return out, nil
}

// Run a fake command!
func (f *fakeHost) Run(cmd string) error {
	_, err := f.CombinedOutput(cmd)
	return err
}

// docker is a fake implementation of docker
func (f *fakeHost) docker(args []string, root bool) (string, error) {
	return "", nil
}

// systemctl is a fake implementation of systemctl
func (f *fakeHost) systemctl(args []string, root bool) (string, error) {
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
		case "start", "restart":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = Running
		case "is-active":
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
			r, err := New(Config{Type: tc.runtime})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			runner := &fakeHost{services: defaultServices}
			err = r.Disable(runner)
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
			"containerd":    Running,
			"crio":          Exited,
			"crio-shutdown": Exited,
		}},
		{"crio", map[string]serviceState{
			"docker":        Exited,
			"docker.socket": Exited,
			"containerd":    Exited,
			"crio":          Running,
			"crio-shutdown": Exited,
		}},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			r, err := New(Config{Type: tc.runtime})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			runner := &fakeHost{services: defaultServices}
			err = r.Enable(runner)
			if err != nil {
				t.Errorf("%s disable unexpected error: %v", tc.runtime, err)
			}
			if diff := cmp.Diff(tc.want, runner.services); diff != "" {
				t.Errorf("service diff (-want +got):\n%s", diff)
			}
		})
	}
}
