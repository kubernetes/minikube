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
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/util/retry"
	"k8s.io/minikube/test/integration/util"
)

func validateRegistryAddon(ctx context.Context, t *testing.T, profile string) {
	MaybeParallel(t)
	mk.MustRun("addons enable registry")
	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("getting kubernetes client: %v", err)
	}
	if err := kapi.WaitForRCToStabilize(client, "kube-system", "registry", time.Minute*5); err != nil {
		t.Fatalf("waiting for registry replicacontroller to stabilize: %v", err)
	}
	rs := labels.SelectorFromSet(labels.Set(map[string]string{"actual-registry": "true"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", rs); err != nil {
		t.Fatalf("waiting for registry pods: %v", err)
	}
	ps := labels.SelectorFromSet(labels.Set(map[string]string{"registry-proxy": "true"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", ps); err != nil {
		t.Fatalf("waiting for registry-proxy pods: %v", err)
	}
	ip, stderr := mk.MustRun("ip")
	ip = strings.TrimSpace(ip)
	endpoint := fmt.Sprintf("http://%s:%d", ip, 5000)
	u, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("failed to parse %q: %v stderr : %s", endpoint, err, stderr)
	}
	t.Log("checking registry access from outside cluster")

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
		t.Fatalf(err.Error())
	}
	t.Log("checking registry access from inside cluster")
	kr := util.NewKubectlRunner(t, profile)
	// TODO: Fix this
	out, _ := kr.RunCommand([]string{
		"run",
		"registry-test",
		"--restart=Never",
		"--image=busybox",
		"-it",
		"--",
		"sh",
		"-c",
		"wget --spider -S 'http://registry.kube-system.svc.cluster.local' 2>&1 | grep 'HTTP/' | awk '{print $2}'"})
	internalCheckOutput := string(out)
	expectedStr := "200"
	if !strings.Contains(internalCheckOutput, expectedStr) {
		t.Errorf("ExpectedStr internalCheckOutput to be: %s. Output was: %s", expectedStr, internalCheckOutput)
	}

	defer func() {
		if _, err := kr.RunCommand([]string{"delete", "pod", "registry-test"}); err != nil {
			t.Errorf("failed to delete pod registry-test")
		}
	}()
	mk.MustRun("addons disable registry")
}
