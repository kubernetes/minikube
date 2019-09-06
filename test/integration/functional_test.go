// +build integration

/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/build/kubernetes/api"
	"k8s.io/apimachinery/pkg/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/minikube/pkg/kapi"
)

// validateFunc are for subtests that share a single setup
type validateFunc func(context.Context, *testing.T, kubernetes.Interface, string)

// TestFunctional are functionality tests which can safely share a profile in parallel
func TestFunctional(t *testing.T) {
	profile := Profile("functional")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile}, StartArgs()...)
	rr, err := RunCmd(ctx, t, Target(), args...)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	client, err := kapi.Client(profile)
	if err != nil {
		t.Fatalf("getting kubernetes client: %v", err)
	}

	t.Run("shared", func(t *testing.T) {
		tests := []struct {
			name           string
			noneCompatible bool
			validator      validateFunc
		}{
			{"AddonManager", true, validateAddonManager},
			{"ComponentHealth", true, validateComponentHealth},
			{"DNS", true, validateDNS},
			{"LogsCmd", true, validateLogsCmd},
			{"KubeContext", true, validateKubeContext},
			{"IngressAddon", false, validateIngressAddon},
			{"MountCmd", false, validateMountCmd},
			{"ProfileCmd", true, validateProfileCmd},
			{"RegistryAddon", true, validateRegistryAddon},
			{"ServicesCmd", true, validateServicesCmd},
			{"PersistentVolumeClaim", true, validatePersistentVolumeClaim},
			{"TunnelCmd", true, validateTunnelCmd},
			{"SSHCmd", false, validateSSHCmd},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				tc.validator(ctx, t, client, profile)
			})
		}
	})
}

// validateAddonManager asserts that the kube-addon-manager pod is deployed properly
func validateAddonManager(ctx context.Context, t *testing.T, client kubernetes.Interface, profile string) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"component": "kube-addon-manager"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "kube-system", selector); err != nil {
		t.Errorf("Error waiting for addon manager to be up")
	}
}

// validateComponentHealth asserts that all Kubernetes components are healthy
func validateComponentHealth(ctx context.Context, t *testing.T, _ kubernetes.Interface, profile string) {
	rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "get", "cs", "-o=json")
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}
	cs := api.ComponentStatusList{}
	d := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes()))
	if err := d.Decode(&cs); err != nil {
		t.Fatalf("decode: %v", err)
	}

	for _, i := range cs.Items {
		status := api.ConditionFalse
		for _, c := range i.Conditions {
			if c.Type != api.ComponentHealthy {
				continue
			}
			status = c.Status
		}
		if status != api.ConditionTrue {
			t.Errorf("unexpected status: %v - item: %+v", status, i)
		}
	}
}

// validateDNS asserts that all Kubernetes DNS is healthy
func validateDNS(ctx context.Context, t *testing.T, client kubernetes.Interface, profile string) {
	rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "busybox.yaml"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	selector := labels.SelectorFromSet(labels.Set(map[string]string{"integration-test": "busybox"}))
	if err := kapi.WaitForPodsWithLabelRunning(client, "default", selector); err != nil {
		t.Errorf("wait failed: %v", err)
	}
	pods, err := client.CoreV1().Pods("default").List(metav1.ListOptions{LabelSelector: "integration-test=busybox"})
	if err != nil {
		t.Errorf("list error: %v", err)
	}
	pod := pods.Items[0].Name

	rr, err = RunCmd(ctx, t, "kubectl", "--context", profile, "exec", pod, "nslookup", "kubernetes.default")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}

	want := []byte("10.96.0.1")
	if !bytes.Contains(rr.Stdout.Bytes(), want) {
		t.Errorf("nslookup: got=%q, want=*%q*", rr.Stdout.Bytes(), want)
	}
}

// validateKubeContext asserts that kubectl config is updated properly
func validateKubeContext(ctx context.Context, t *testing.T, _ kubernetes.Interface, profile string) {
	rr, err := RunCmd(ctx, t, "kubectl", "config", "current-context")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if !strings.Contains(rr.Stdout.String(), profile) {
		t.Errorf("current-context = %q, want %q", rr.Stdout.String(), profile)
	}
}

// validateConfigCmd asserts basic "config" command functionality
func validateConfigCmd(ctx context.Context, t *testing.T, _ kubernetes.Interface, profile string) {
	tests := []struct {
		args    []string
		wantOut string
		wantErr string
	}{
		{[]string{"unset", "cpus"}, "", ""},
		{[]string{"get", "cpus"}, "", "Error: specified key could not be found in config"},
		{[]string{"set", "cpus 2"}, "! These changes will take effect upon a minikube delete and then a minikube start", ""},
		{[]string{"get", "cpus"}, "2", ""},
		{[]string{"unset", "cpus"}, "", ""},
		{[]string{"get", "cpus"}, "", "Error: specified key could not be found in config"},
	}

	for _, tc := range tests {
		args := append([]string{"-p", profile, "config"}, tc.args...)
		rr, err := RunCmd(ctx, t, Target(), args...)
		if err != nil {
			t.Errorf("%s failed: %v", rr.Args, err)
		}

		got := strings.TrimSpace(rr.Stdout.String())
		if got != tc.wantOut {
			t.Errorf("config %s stdout got: %q, want: %q", tc.args, got, tc.wantOut)
		}
		got = strings.TrimSpace(rr.Stderr.String())
		if got != tc.wantErr {
			t.Errorf("config %s stderr got: %q, want: %q", tc.args, got, tc.wantErr)
		}
	}
}

// validateLogsCmd asserts basic "logs" command functionality
func validateLogsCmd(ctx context.Context, t *testing.T, _ kubernetes.Interface, profile string) {
	rr, err := RunCmd(ctx, t, Target(), "-p", profile, "logs")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	for _, word := range []string{"Docker", "apiserver", "Linux", "kubelet"} {
		if !strings.Contains(rr.Stdout.String(), word) {
			t.Errorf("minikube logs missing expected word: %q", word)
		}
	}
}

// validateProfileCmd asserts basic "profile" command functionality
func validateProfileCmd(ctx context.Context, t *testing.T, _ kubernetes.Interface, profile string) {
	rr, err := RunCmd(ctx, t, Target(), "profile", "list")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}

// validateServiceCmd asserts basic "service" command functionality
func validateServicesCmd(ctx context.Context, t *testing.T, _ kubernetes.Interface, profile string) {
	rr, err := RunCmd(ctx, t, Target(), "-p", profile, "service", "list")
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if !strings.Contains(rr.Stdout.String(), "kubernetes") {
		t.Errorf("service list got %q, wanted *kubernetes*", rr.Stdout.String())
	}
}

// validateSSHCmd asserts basic "ssh" command functionality
func validateSSHCmd(ctx context.Context, t *testing.T, _ kubernetes.Interface, profile string) {
	want := "hello\r\n"
	rr, err := RunCmd(ctx, t, Target(), "-p", profile, "ssh", fmt.Sprintf("echo hello"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
	if rr.Stdout.String() != want {
		t.Errorf("%v = %q, want = %q", rr.Args, rr.Stdout.String(), want)
	}
}
