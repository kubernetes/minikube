package runtime

import (
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
		t.Run(tc.input, func(t *testing.T) {
			r, err := New(Config{Type: tc.runtime})
			if err != nil {
				t.Fatalf("New(%s): %v", tc.runtime, err)
			}
			got := r.Name(tc.input)
			if got != tc.want {
				t.Errorf("Name(%s) = %q, want: %q", tc.input, got, tc.want)
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
		{"empty", "", map[string]string{}, map[string]string{"container-runtime": ""}},
		{"crio", "crio", map[string]string{}, map[string]string{
			"container-runtime":          "crio",
			"container-runtime-endpoint": "/var/run/crio/crio.sock",
			"image-service-endpoint":     "/var/run/crio/crio.sock",
			"runtime-request-timeout":    "15m",
		}},
		{"crio-in-map", "", map[string]string{"container-runtime": "crio"}, map[string]string{
			"container-runtime": "crio",
		}},
		{"containerd", "containerd", map[string]string{}, map[string]string{
			"container-runtime":          "containerd",
			"container-runtime-endpoint": "unix:///run/containerd/containerd.sock",
			"image-service-endpoint":     "unix:///run/containerd/containerd.sock",
			"runtime-request-timeout":    "15m",
		}},
		{"conflicting-runtimes", "crio", map[string]string{"container-runtime": "containerd"}, map[string]string{
			"container-runtime": "containerd",
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
