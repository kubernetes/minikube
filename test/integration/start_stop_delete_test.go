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
	"net"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"k8s.io/minikube/test/integration/util"
)

func TestStartStop(t *testing.T) {
	tests := []struct {
		runtime string
	}{
		{runtime: "docker"},
		{runtime: "containerd"},
		{runtime: "crio"},
	}

	for _, test := range tests {
		t.Run(test.runtime, func(t *testing.T) {
			runner := NewMinikubeRunner(t)
			if test.runtime != "" && usingNoneDriver(runner) {
				t.Skipf("skipping, can't use %s with none driver", test.runtime)
			}

			runner.RunCommand("config set WantReportErrorPrompt false", true)
			runner.RunCommand("delete", false)
			runner.CheckStatus(state.None.String())

			runner.SetRuntime(test.runtime)
			runner.Start()
			runner.CheckStatus(state.Running.String())

			ip := runner.RunCommand("ip", true)
			ip = strings.TrimRight(ip, "\n")
			if net.ParseIP(ip) == nil {
				t.Fatalf("IP command returned an invalid address: %s", ip)
			}

			checkStop := func() error {
				runner.RunCommand("stop", true)
				return runner.CheckStatusNoFail(state.Stopped.String())
			}

			if err := util.Retry(t, checkStop, 5*time.Second, 6); err != nil {
				t.Fatalf("timed out while checking stopped status: %v", err)
			}

			runner.Start()
			runner.CheckStatus(state.Running.String())

			runner.RunCommand("delete", true)
			runner.CheckStatus(state.None.String())
		})
	}
}
