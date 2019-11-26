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
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/constants"
)

func TestStartStop(t *testing.T) {
	MaybeParallel(t)

	t.Run("group", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
		}{
			{"docker", []string{
				"--cache-images=false",
				fmt.Sprintf("--kubernetes-version=%s", constants.OldestKubernetesVersion),
				// default is the network created by libvirt, if we change the name minikube won't boot
				// because the given network doesn't exist
				"--kvm-network=default",
				"--kvm-qemu-uri=qemu:///system",
				"--disable-driver-mounts",
				"--keep-context=false",
				"--container-runtime=docker",
			}},
			{"cni", []string{
				"--feature-gates",
				"ServerSideApply=true",
				"--network-plugin=cni",
				"--extra-config=kubelet.network-plugin=cni",
				"--extra-config=kubeadm.pod-network-cidr=192.168.111.111/16",
				fmt.Sprintf("--kubernetes-version=%s", constants.NewestKubernetesVersion),
			}},
			{"containerd", []string{
				"--container-runtime=containerd",
				"--docker-opt",
				"containerd=/var/run/containerd/containerd.sock",
				"--apiserver-port=8444",
			}},
			{"crio", []string{
				"--container-runtime=crio",
				"--disable-driver-mounts",
				"--extra-config=kubeadm.ignore-preflight-errors=SystemVerification",
			}},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				WaitForStartSlot(t)

				if !strings.Contains(tc.name, "docker") && NoneDriver() {
					t.Skipf("skipping %s - incompatible with none driver", t.Name())
				}

				profile := UniqueProfileName(tc.name)
				ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
				defer CleanupWithLogs(t, profile, cancel)

				startArgs := append([]string{"start", "-p", profile, "--alsologtostderr", "-v=3", "--wait=true"}, tc.args...)
				startArgs = append(startArgs, StartArgs()...)
				rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
				if err != nil {
					// Fatal so that we may collect logs before stop/delete steps
					t.Fatalf("%s failed: %v", rr.Args, err)
				}

				// SADNESS: 0/1 nodes are available: 1 node(s) had taints that the pod didn't tolerate.
				if strings.Contains(tc.name, "cni") {
					t.Logf("WARNING: cni mode requires additional setup before pods can schedule :(")
				} else {
					// schedule a pod to assert persistence
					rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "busybox.yaml")))
					if err != nil {
						t.Fatalf("%s failed: %v", rr.Args, err)
					}

					names, err := PodWait(ctx, t, profile, "default", "integration-test=busybox", 4*time.Minute)
					if err != nil {
						t.Fatalf("wait: %v", err)
					}

					// Use this pod to confirm that the runtime resource limits are sane
					rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", names[0], "--", "/bin/sh", "-c", "ulimit -n"))
					if err != nil {
						t.Fatalf("ulimit: %v", err)
					}

					got, err := strconv.ParseInt(strings.TrimSpace(rr.Stdout.String()), 10, 64)
					if err != nil {
						t.Errorf("ParseInt(%q): %v", rr.Stdout.String(), err)
					}

					// Arbitrary value set by some container runtimes. If higher, apps like MySQL may make bad decisions.
					expected := int64(1048576)
					if got != expected {
						t.Errorf("'ulimit -n' returned %d, expected %d", got, expected)
					}
				}

				rr, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile, "--alsologtostderr", "-v=3"))
				if err != nil {
					t.Errorf("%s failed: %v", rr.Args, err)
				}

				got := Status(ctx, t, Target(), profile)
				if got != state.Stopped.String() {
					t.Errorf("status = %q; want = %q", got, state.Stopped)
				}

				WaitForStartSlot(t)
				rr, err = Run(t, exec.CommandContext(ctx, Target(), startArgs...))
				if err != nil {
					// Explicit fatal so that failures don't move directly to deletion
					t.Fatalf("%s failed: %v", rr.Args, err)
				}

				if strings.Contains(tc.name, "cni") {
					t.Logf("WARNING: cni mode requires additional setup before pods can schedule :(")
				} else if _, err := PodWait(ctx, t, profile, "default", "integration-test=busybox", 2*time.Minute); err != nil {
					t.Fatalf("wait: %v", err)
				}

				got = Status(ctx, t, Target(), profile)
				if got != state.Running.String() {
					t.Errorf("status = %q; want = %q", got, state.Running)
				}

				if !*cleanup {
					// Normally handled by cleanuprofile, but not fatal there
					rr, err = Run(t, exec.CommandContext(ctx, Target(), "delete", "-p", profile))
					if err != nil {
						t.Errorf("%s failed: %v", rr.Args, err)
					}
				}
			})
		}
	})
}
