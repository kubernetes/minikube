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
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/constants"
)

func TestStartStop(t *testing.T) {
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
				"--wait=false",
				"--disable-driver-mounts",
				"--keep-context=false",
			}},
			{"cni", []string{
				"--feature-gates",
				"ServerSideApply=true",
				"--network-plugin=cni",
				"--extra-config=kubelet.network-plugin=cni",
				"--extra-config=kubeadm.pod-network-cidr=192.168.111.111/16",
				fmt.Sprintf("--kubernetes-version=%s", constants.NewestKubernetesVersion),
				"--wait=false",
			}},
			{"containerd", []string{
				"--container-runtime=containerd",
				"--docker-opt",
				"containerd=/var/run/containerd/containerd.sock",
				"--apiserver-port=8444",
				"--wait=false",
			}},
			{"crio", []string{
				"--container-runtime=crio",
				"--disable-driver-mounts",
				"--extra-config=kubeadm.ignore-preflight-errors=SystemVerification",
				"--wait=false",
			}},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)

				if !strings.Contains(tc.name, "docker") && NoneDriver() {
					t.Skipf("skipping %s - incompatible with none driver", t.Name())
				}

				profile := Profile(tc.name)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
				defer CleanupWithLogs(t, profile, cancel)

				// Use copious amounts of debugging for this stress test: we will need it.
				startArgs := append([]string{"start", "-p", profile, "--alsologtostderr", "-v=8"}, tc.args...)
				rr, err := Run(ctx, t, Target(), startArgs...)
				if err != nil {
					t.Errorf("%s failed: %v", rr.Args, err)
				}

				rr, err = Run(ctx, t, Target(), "stop", "-p", profile)
				if err != nil {
					t.Errorf("%s failed: %v", rr.Args, err)
				}

				got := Status(ctx, t, Target(), profile)
				if got != state.Stopped.String() {
					t.Errorf("status = %q; want = %q", got, state.Stopped)
				}

				rr, err = Run(ctx, t, Target(), startArgs...)
				if err != nil {
					// Explicit fatal so that failures don't move directly to deletion
					t.Fatalf("%s failed: %v", rr.Args, err)
				}

				got = Status(ctx, t, Target(), profile)
				if got != state.Running.String() {
					t.Errorf("status = %q; want = %q", got, state.Running)
				}

				// Normally handled by cleanuprofile, but not fatal there
				rr, err = Run(ctx, t, Target(), "delete", "-p", profile)
				if err != nil {
					t.Errorf("%s failed: %v", rr.Args, err)
				}
			})
		}
	})
}
