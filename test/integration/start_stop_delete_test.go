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
	"encoding/json"
	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/test/integration/util"
	"net"
	"strings"
	"testing"
	"time"
)

func TestStartStop(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		assertCustom func(t *testing.T)
	}{
		{"docker+cache", []string{"--container-runtime=docker", "--cache-images"}, nil},
		{"containerd+cache", []string{"--container-runtime=containerd", "--docker-opt containerd=/var/run/containerd/containerd.sock", "--cache-images"}, nil},
		{"crio+cache", []string{"--container-runtime=crio", "--cache-images"}, nil},
		{"podCidr", []string{"--pod-network-cidr=192.168.111.111/16"}, assertPodCIDR},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := NewMinikubeRunner(t)
			if !strings.Contains(test.name, "docker") && usingNoneDriver(r) {
				t.Skipf("skipping %s - incompatible with none driver", test.name)
			}

			r.RunCommand("config set WantReportErrorPrompt false", true)
			r.RunCommand("delete", false)
			r.CheckStatus(state.None.String())
			r.Start(test.args...)
			r.CheckStatus(state.Running.String())

			if test.assertCustom != nil {
				test.assertCustom(t)
			}

			ip := r.RunCommand("ip", true)
			ip = strings.TrimRight(ip, "\n")
			if net.ParseIP(ip) == nil {
				t.Fatalf("IP command returned an invalid address: %s", ip)
			}

			checkStop := func() error {
				r.RunCommand("stop", true)
				return r.CheckStatusNoFail(state.Stopped.String())
			}

			if err := util.Retry(t, checkStop, 5*time.Second, 6); err != nil {
				t.Fatalf("timed out while checking stopped status: %v", err)
			}

			r.Start(test.args...)
			r.CheckStatus(state.Running.String())

			r.RunCommand("delete", true)
			r.CheckStatus(state.None.String())
		})
	}
}

func assertPodCIDR(t *testing.T) {
	kr := util.NewKubectlRunner(t)
	out, err := kr.RunCommand([]string{"get", "nodes", "-o", "json"})
	if err != nil {
		t.Fatalf("Failed to obtain nodes info")
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(out), &result)

	items := result["items"].([]interface{})
	for _, item := range items {
		spec := item.(map[string]interface{})["spec"]
		podCidr := spec.(map[string]interface{})["podCIDR"].(string)

		if !strings.HasPrefix(podCidr, "192.168.0.0") {
			t.Errorf("Unexpected podCIDR: %s", podCidr)
		}
	}
}
