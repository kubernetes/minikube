package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/util/retry"
)

func validateRegistryAddon(ctx context.Context, t *testing.T, client kubernetes.Interface, profile string) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer func() {
		cancel()
		rr, err := RunCmd(context.Background(), t, Target(), "addons", "disable", "registry")
		if err != nil {
			t.Logf("cleanup failed: %s: %v (probably ok)", rr.Args, err)
		}
	}()

	rr, err := RunCmd(ctx, t, Target(), "addons", "enable", "registry")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if err := kapi.WaitForRCToStabilize(client, "kube-system", "registry", time.Minute*5); err != nil {
		t.Errorf("waiting for registry replicacontroller to stabilize: %v", err)
	}
	rs := labels.SelectorFromSet(labels.Set(map[string]string{"actual-registry": "true"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", rs); err != nil {
		t.Errorf("waiting for registry pods: %v", err)
	}
	ps := labels.SelectorFromSet(labels.Set(map[string]string{"registry-proxy": "true"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", ps); err != nil {
		t.Errorf("waiting for registry-proxy pods: %v", err)
	}

	rr, err = RunCmd(ctx, t, Target(), "-p", profile, "ip")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if rr.Stderr.String() != "" {
		t.Errorf("%s: unexpected stderr: %s", rr.Args, rr.Stderr)
	}
	endpoint := fmt.Sprintf("http://%s:%d", strings.TrimSpace(rr.Stdout.String()), 5000)
	u, err := url.Parse(endpoint)
	if err != nil {
		t.Errorf("failed to parse %q: %v", endpoint, err)
	}

	// Check access from outside the cluster on port 5000, validing connectivity via registry-proxy
	checkExternalAccess := func() error {
		resp, err := retryablehttp.Get(u.String())
		if err != nil {
			t.Errorf("failed get: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s returned status code %d, expected %d.\n", u, resp.StatusCode, http.StatusOK)
		}
		return nil
	}

	if err := retry.Expo(checkExternalAccess, 500*time.Millisecond, 2*time.Minute); err != nil {
		t.Errorf(err.Error())
	}
	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "run", "registry-test", "--image=busybox", "-it", "--", "curl", "-vvv", "http://registry.kube-system.svc.cluster.local")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	want := "HTTP/1.1 200"
	if !strings.Contains(rr.Stdout.String(), want) {
		t.Errorf("curl = %q, want *%s*", rr.Stdout.String(), want)
	}
	rr, err = RunCmd(ctx, t, Target(), "addons", "disable", "registry")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}
