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
	Restarted
)

// FakeRunner is a command runner that isn't very smart.
type FakeRunner struct {
	cmds     []string
	services map[string]serviceState
	t        *testing.T
}

// NewFakeRunner returns a CommandRunner which emulates a systemd host
func NewFakeRunner(t *testing.T) *FakeRunner {
	return &FakeRunner{
		services: map[string]serviceState{},
		cmds:     []string{},
		t:        t,
	}
}

// Run a fake command!
func (f *FakeRunner) CombinedOutput(cmd string) (string, error) {
	f.cmds = append(f.cmds, cmd)
	out := ""

	root := false
	args := strings.Split(cmd, " ")
	bin, args := args[0], args[1:]
	f.t.Logf("bin=%s args=%v", bin, args)
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
func (f *FakeRunner) Run(cmd string) error {
	_, err := f.CombinedOutput(cmd)
	return err
}

// docker is a fake implementation of docker
func (f *FakeRunner) docker(args []string, root bool) (string, error) {
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
			f.t.Logf("stopped %s", svc)
		case "start":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = Running
			f.t.Logf("started %s", svc)
		case "restart":
			if !root {
				return out, fmt.Errorf("not root")
			}
			f.services[svc] = Restarted
			f.t.Logf("restarted %s", svc)
		case "is-active":
			f.t.Logf("%s is-status: %v", svc, state)
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
			"docker":        Restarted,
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
