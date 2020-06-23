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

	"k8s.io/minikube/pkg/kapi"
)

func TestNetworkPlugins(t *testing.T) {
	MaybeParallel(t)

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
			{"calico", []string{"--cni=calico"}, "cni", "app=calico", true},
			// kindnet only configures hairpin properly for the Docker driver
			{"kindnet", []string{"--cni=kindnet"}, "cni", "app=kindnet", DockerDriver()},
			{"false", []string{"--cni=false"}, "", "", true},
			{"custom-weave", []string{fmt.Sprintf("--cni=%s", filepath.Join(*testdataDir, "netcat-deployment.yaml"))}, "cni", "name=weave-net", true},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				profile := UniqueProfileName(tc.name)
				ctx, cancel := context.WithTimeout(context.Background(), Minutes(20))
				defer Cleanup(t, profile, cancel)

				startArgs := append([]string{"start", "-p", profile, "--memory=1500", "--alsologtostderr", "--wait=true"}, tc.args...)
				startArgs = append(startArgs, StartArgs()...)

				t.Run("Start", func(t *testing.T) {
					_, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
					if err != nil {
						t.Fatalf("failed start: %v", err)
					}
				})

				if tc.podLabel != "" {
					t.Run("ControllerPod", func(t *testing.T) {
						if _, err := PodWait(ctx, t, profile, "kube-system", tc.podLabel, Minutes(4)); err != nil {
							t.Fatalf("failed waiting for %s labeled pod: %v", tc.podLabel, err)
						}
					})
				}
				t.Run("KubeletFlags", func(t *testing.T) {
					rr, err := Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "pgrep -a kubelet"))
					if err != nil {
						t.Fatalf("ssh failed: %v", err)
					}
					out := rr.Stdout.String()

					if tc.kubeletPlugin == "" {
						if strings.Contains(out, "--network-plugin") {
							t.Errorf("expected no network plug-in, got %s", out)
						}
					} else {
						if !strings.Contains(out, fmt.Sprintf("--network-plugin=%s", tc.kubeletPlugin)) {
							t.Errorf("expected --network-plugin=%s, got %s", tc.kubeletPlugin, out)
						}
					}

				})

				t.Run("NetCatPod", func(t *testing.T) {
					_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "netcat-deployment.yaml")))
					if err != nil {
						t.Errorf("failed to apply netcat manifest: %v", err)
					}

					client, err := kapi.Client(profile)
					if err != nil {
						t.Fatalf("failed to get Kubernetes client for %s: %v", profile, err)
					}

					if err := kapi.WaitForDeploymentToStabilize(client, "default", "netcat", Minutes(4)); err != nil {
						t.Errorf("failed waiting for netcat deployment to stabilize: %v", err)
					}

					if _, err := PodWait(ctx, t, profile, "default", "app=netcat", Minutes(4)); err != nil {
						t.Fatalf("failed waiting for netcat pod: %v", err)
					}

				})

				t.Run("DNS", func(t *testing.T) {
					rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "nslookup", "kubernetes.default"))
					if err != nil {
						t.Errorf("failed to do nslookup on kubernetes.default: %v", err)
					}

					want := []byte("10.96.0.1")
					if !bytes.Contains(rr.Stdout.Bytes(), want) {
						t.Errorf("failed nslookup: got=%q, want=*%q*", rr.Stdout.Bytes(), want)
					}
				})

				t.Run("Localhost", func(t *testing.T) {
					_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "/bin/sh", "-c", "nc -w 1 -i 1 -z localhost 8080"))
					if err != nil {
						t.Errorf("localhost connection failed: %v", err)
					}
				})

				t.Run("HairPin", func(t *testing.T) {
					_, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", "deployment/netcat", "--", "/bin/sh", "-c", "nc -w 1 -i 1 -z netcat 8080"))
					if tc.hairpin && err != nil {
						t.Fatalf("hairpin connaction failed: %v", err)
					}

					if !tc.hairpin && err == nil {
						t.Fatalf("hairpin connection unexpectedly succeeded - misconfigured test?")
					}
				})
			})
		}
	})
}
