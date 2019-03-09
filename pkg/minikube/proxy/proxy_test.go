package proxy

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/minikube/config"
)

func TestEnvironment(t *testing.T) {
	want := "moo"
	os.Setenv("NO_PROXY", want)
	got := Environment()["bypass"]
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
func TestRecommended(t *testing.T) {
	c := config.KubernetesConfig{NodeIP: "1.2.3.4"}
	os.Setenv("ALL_PROXY", "moo")
	got := Recommended(c)
	want := map[string]string{
		"HTTP_PROXY":  "moo",
		"HTTPS_PROXY": "moo",
		"NO_PROXY":    "1.2.3.4,127.0.0.1,moo",
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff: %s", diff)

	}
}
func TestNoProxyWithValue(t *testing.T) {
	var tests = []struct {
		env  string
		want string
	}{
		{"", "1.2.3.4/8,127.0.0.1"},
		{"127.0.0.1", "1.2.3.4/8,127.0.0.1"},
		{"x.y.com", "1.2.3.4/8,127.0.0.1,x.y.com"},
	}
	c := config.KubernetesConfig{ServiceCIDR: "1.2.3.4/8"}

	for _, tc := range tests {
		t.Run(tc.env, func(t *testing.T) {
			os.Setenv("NO_PROXY", tc.env)
			got := NoProxy(c)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
