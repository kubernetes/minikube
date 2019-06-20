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
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/test/integration/util"
)

func TestStartStop(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"nocache_oldest", []string{
			"--cache-images=false",
			fmt.Sprintf("--kubernetes-version=%s", constants.OldestKubernetesVersion),
			// default is the network created by libvirt, if we change the name minikube won't boot
			// because the given network doesn't exist
			"--kvm-network=default",
		}},
		{"feature_gates_newest_cni", []string{
			"--feature-gates",
			"ServerSideApply=true",
			"--network-plugin=cni",
			"--extra-config=kubelet.network-plugin=cni",
			"--extra-config=kubeadm.pod-network-cidr=192.168.111.111/16",
			fmt.Sprintf("--kubernetes-version=%s", constants.NewestKubernetesVersion),
		}},
		{"containerd_and_non_default_apiserver_port", []string{
			"--container-runtime=containerd",
			"--docker-opt containerd=/var/run/containerd/containerd.sock",
			"--apiserver-port=8444",
		}},
		{"crio_ignore_preflights", []string{
			"--container-runtime=crio",
			"--extra-config",
			"kubeadm.ignore-preflight-errors=SystemVerification",
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewMinikubeRunner(t)
			if !strings.Contains(test.name, "docker") && usingNoneDriver(r) {
				t.Skipf("skipping %s - incompatible with none driver", test.name)
			}

			kctl := util.NewKubectlRunner(t)
			r.RunCommand("config set WantReportErrorPrompt false", true)
			r.RunCommand("delete", false)
			r.CheckStatus(state.None.String())
			r.Start(test.args...)
			r.CheckStatus(state.Running.String())

			ip := r.RunCommand("ip", true)
			ip = strings.TrimRight(ip, "\n")
			if net.ParseIP(ip) == nil {
				t.Fatalf("IP command returned an invalid address: %s", ip)
			}

			// check for the current-context before and after the stop
			currCtx := kctl.CurrentContext()
			if currCtx != "minikube" {
				t.Fatalf("Got current-context - %q, want %q", currCtx, "minikube")
			}
			checkStop := func() error {
				r.RunCommand("stop", false)
				err := r.CheckStatusNoFail(state.Stopped.String())
				if err == nil {
					afterStopCtx := kctl.CurrentContext()
					if afterStopCtx == "" {
						return nil
					}
					return fmt.Errorf("Got current-context - %q, want %q", afterStopCtx, "")
				}
				return err
			}

			if err := util.Retry(t, checkStop, 20*time.Second, 10); err != nil {
				t.Fatalf("Timed checking minikube stop : %v", err)
			}

			r.Start(test.args...)
			r.CheckStatus(state.Running.String())

			r.RunCommand("delete", true)
			r.CheckStatus(state.None.String())
		})
	}
}
