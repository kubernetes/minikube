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
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/util/retry"
)

func TestNetworkPlugins(t *testing.T) {
	MaybeParallel(t)
	if NoneDriver() {
		t.Skip("skipping since test for none driver")
	}

	t.Run("group", func(t *testing.T) {
		tests := []struct {
			name          string
			args          []string
			kubeletPlugin string
			podLabel      string
			hairpin       bool
		}{
			{"auto", []string{}, "", "", false},
			{"kubenet", []string{"--network-plugin=kubenet"}, "kubenet", "", true},
			{"bridge", []string{"--cni=bridge"}, "cni", "", true},
			{"enable-default-cni", []string{"--enable-default-cni=true"}, "cni", "", true},
			{"flannel", []string{"--cni=flannel"}, "cni", "app=flannel", true},
			{"kindnet", []string{"--cni=kindnet"}, "cni", "app=kindnet", true},
			{"false", []string{"--cni=false"}, "", "", false},
			{"custom-weave", []string{fmt.Sprintf("--cni=%s", filepath.Join(*testdataDir, "weavenet.yaml"))}, "cni", "", true},
			{"calico", []string{"--cni=calico"}, "cni", "k8s-app=calico-node", true},
			{"cilium", []string{"--cni=cilium"}, "cni", "k8s-app=cilium", true},
		}

		for _, tc := range tests {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				if DockerDriver() && strings.Contains(tc.name, "flannel") {
					t.Skipf("flannel is not yet compatible with Docker driver: iptables v1.8.3 (legacy): Couldn't load target `CNI-x': No such file or directory")
				}

				start := time.Now()
				MaybeParallel(t)
				profile := UniqueProfileName(tc.name)

				ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
				defer CleanupWithLogs(t, profile, cancel)

				startArgs := append([]string{"start", "-p", profile, "--memory=1800", "--alsologtostderr", "--wait=true", "--wait-timeout=5m"}, tc.args...)
				startArgs = append(startArgs, StartArgs()...)

				t.Run("Start", func(t *testing.T) {
					_, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
					if err != nil {
						t.Fatalf("failed start: %v", err)
					}
				})

				if !t.Failed() && tc.podLabel != "" {
					t.Run("ControllerPod", func(t *testing.T) {
						if _, err := PodWait(ctx, t, profile, "kube-system", tc.podLabel, Minutes(10)); err != nil {
							t.Fatalf("failed waiting for %s labeled pod: %v", tc.podLabel, err)
						}
					})
				}
				if !t.Failed() {
					t.Run("KubeletFlags", func(t *testing.T) {
						// none does not support 'minikube ssh'
						rr, err := Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "pgrep -a kubelet"))
						if NoneDriver() {
							rr, err = Run(t, exec.CommandContext(ctx, "pgrep", "-a", "kubelet"))
						}
						if err != nil {
							t.Fatalf("ssh failed: %v", err)
						}
						out := rr.Stdout.String()
						verifyKubeletFlagsOutput(t, tc.kubeletPlugin, out)
					})
				}

				if !t.Failed() {
					t.Run("NetCatPod", func(t *testing.T) {
						_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "netcat-deployment.yaml")))
						if err != nil {
							t.Errorf("failed to apply netcat manifest: %v", err)
						}

						client, err := kapi.Client(profile)
						if err != nil {
							t.Fatalf("failed to get Kubernetes client for %s: %v", profile, err)
						}

						if err := kapi.WaitForDeploymentToStabilize(client, "default", "netcat", Minutes(15)); err != nil {
							t.Errorf("failed waiting for netcat deployment to stabilize: %v", err)
						}

						if _, err := PodWait(ctx, t, profile, "default", "app=netcat", Minutes(15)); err != nil {
							t.Fatalf("failed waiting for netcat pod: %v", err)
						}

					})
				}

				if strings.Contains(tc.name, "weave") {
					t.Skipf("skipping remaining tests for weave, as results can be unpredictable")
				}

				if !t.Failed() {
					t.Run("DNS", func(t *testing.T) {
						var rr *RunResult
						var err error

						nslookup := func() error {
							rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nslookup", "kubernetes.default"))
							return err
						}

						// If the coredns process was stable, this retry wouldn't be necessary.
						if err := retry.Expo(nslookup, 1*time.Second, Minutes(6)); err != nil {
							t.Errorf("failed to do nslookup on kubernetes.default: %v", err)
						}

						want := []byte("10.96.0.1")
						if !bytes.Contains(rr.Stdout.Bytes(), want) {
							t.Errorf("failed nslookup: got=%q, want=*%q*", rr.Stdout.Bytes(), want)
						}
					})
				}

				if !t.Failed() {
					t.Run("Localhost", func(t *testing.T) {
						tryLocal := func() error {
							_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "/bin/sh", "-c", "nc -w 5 -i 5 -z localhost 8080"))
							return err
						}

						if err := retry.Expo(tryLocal, 1*time.Second, Seconds(60)); err != nil {
							t.Errorf("failed to connect via localhost: %v", err)
						}
					})
				}

				if !t.Failed() {
					t.Run("HairPin", func(t *testing.T) {
						tryHairPin := func() error {
							_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "/bin/sh", "-c", "nc -w 5 -i 5 -z netcat 8080"))
							return err
						}

						if tc.hairpin {
							if err := retry.Expo(tryHairPin, 1*time.Second, Seconds(60)); err != nil {
								t.Errorf("failed to connect via pod host: %v", err)
							}
						} else {
							if tryHairPin() == nil {
								t.Fatalf("hairpin connection unexpectedly succeeded - misconfigured test?")
							}
						}
					})
				}

				t.Logf("%q test finished in %s, failed=%v", tc.name, time.Since(start), t.Failed())
			})
		}
	})
}

func verifyKubeletFlagsOutput(t *testing.T, kubeletPlugin, out string) {
	if kubeletPlugin == "" {
		if strings.Contains(out, "--network-plugin") && ContainerRuntime() == "docker" {
			t.Errorf("expected no network plug-in, got %s", out)
		}
		if !strings.Contains(out, "--network-plugin=cni") && ContainerRuntime() != "docker" {
			t.Errorf("expected cni network plugin with conatinerd/crio, got %s", out)
		}
	} else if !strings.Contains(out, fmt.Sprintf("--network-plugin=%s", kubeletPlugin)) {
		t.Errorf("expected --network-plugin=%s, got %s", kubeletPlugin, out)
	}
}
