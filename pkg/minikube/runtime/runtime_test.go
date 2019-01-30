package runtime

import (
	"fmt"
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
		name    string
		runtime string
		cfg     map[string]string
		want    map[string]string
	}{
		{"empty", "", map[string]string{}, map[string]string{"container-runtime": "docker"}},
		{"crio", "crio", map[string]string{}, map[string]string{
			"container-runtime":          "remote",
			"container-runtime-endpoint": "/var/run/crio/crio.sock",
			"image-service-endpoint":     "/var/run/crio/crio.sock",
			"runtime-request-timeout":    "15m",
		}},
		{"containerd", "containerd", map[string]string{}, map[string]string{
			"container-runtime":          "remote",
			"container-runtime-endpoint": "unix:///run/containerd/containerd.sock",
			"image-service-endpoint":     "unix:///run/containerd/containerd.sock",
			"runtime-request-timeout":    "15m",
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := New(Config{Type: tc.runtime})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}

			got := r.KubeletOptions(tc.cfg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("KubeletOptions(%s, %v) returned diff (-want +got):\n%s", tc.runtime, tc.cfg, diff)
			}
		})
	}
}

// fakeHost is a command runner that isn't very smart.
type fakeHost struct {
	cmds []string
}

// Run the fake command!
func (f *fakeHost) Run(cmd string) error {
	f.cmds = append(f.cmds, cmd)
	return fmt.Errorf("unknown command: %s", cmd)
}

func TestEnable(t *testing.T) {
	var tests = []struct {
		runtime string
		want    string
	}{
		{"docker", "Docker"},
		{"crio", "CRIO"},
		{"containerd", "containerd"},
	}
	for _, tc := range tests {
		t.Run(tc.runtime, func(t *testing.T) {
			r, err := New(Config{Type: tc.runtime})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			runner := &fakeHost{}
			err = r.Disable(runner)
			if err != nil {
				t.Errorf("%s disable unexpected error: %v", tc.runtime, err)
			}
		})
	}
}
