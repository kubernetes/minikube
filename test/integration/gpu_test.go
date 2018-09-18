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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	api "k8s.io/api/core/v1"
	"k8s.io/minikube/test/integration/util"
)

func TestGpu(t *testing.T) {

	runner := NewMinikubeRunner(t)
	runner.RunCommand("config set WantReportErrorPrompt false", true)
	runner.RunCommand("delete", false)
	runner.CheckStatus(state.None.String())

	// Check that we can bring up VM with --gpu flag.
	runner.Start()
	runner.CheckStatus(state.Running.String())

	ip := runner.RunCommand("ip", true)
	ip = strings.TrimRight(ip, "\n")
	if net.ParseIP(ip) == nil {
		t.Fatalf("IP command returned an invalid address: %s", ip)
	}

	// Enable the addons that install the NVIDIA driver and expose the GPU resources.
	runner.RunCommand("addons enable nvidia-driver-installer", true)
	runner.RunCommand("addons enable nvidia-gpu-device-plugin", true)

	// Wait for the device plugin and driver installer to become running.
	if err := util.WaitForNvidiaDriverInstallerRunning(t); err != nil {
		t.Errorf("waiting for nvidia-driver-installer to be up: %s", err)
	}
	if err := util.WaitForNvidiaGpuDevicePluginRunning(t); err != nil {
		t.Errorf("waiting for nvidia-gpu-device-plugin to be up: %s", err)
	}

	kubectlRunner := util.NewKubectlRunner(t)

	// Wait for the 'nvidia.com/gpu' capacity to show up in the nodes.
	// If this completes successfully, we can be sure that driver installation was successful.
	checkCapacity := func() error {
		nodes := api.NodeList{}
		if err := kubectlRunner.RunCommandParseOutput([]string{"get", "nodes"}, &nodes); err != nil {
			return fmt.Errorf("parsing 'kubectl get nodes' output")
		}
		for _, node := range nodes.Items {
			if _, ok := node.Status.Capacity["nvidia.com/gpu"]; !ok {
				return fmt.Errorf("nvidia.com/gpu not present in capacity.")
			}
		}
		return nil
	}
	if err := util.Retry(t, checkCapacity, 10*time.Second, 30); err != nil {
		t.Errorf("timed out waiting for driver installation to finish: %v", err)
	}

	// Make sure that CUDA workload can run.
	workloadPath, err := filepath.Abs("testdata/cuda.yaml")
	if err != nil {
		t.Errorf("constructing path to cuda workload manifest: %v", err)
	}
	if _, err := kubectlRunner.RunCommand([]string{"create", "-f", workloadPath}); err != nil {
		t.Errorf("creating cuda workload: %s", err)
	}
	if err := util.WaitForCudaSuccess(t); err != nil {
		t.Errorf("waiting for cuda-vector-add to be done: %s", err)
	}

	// Check that stopping and restarting works when using --gpu.
	checkStop := func() error {
		runner.RunCommand("stop", true)
		return runner.CheckStatusNoFail(state.Stopped.String())
	}
	if err := util.Retry(t, checkStop, 5*time.Second, 6); err != nil {
		t.Fatalf("timed out while checking stopped status: %s", err)
	}

	runner.Start()
	runner.CheckStatus(state.Running.String())

	runner.RunCommand("delete", true)
	runner.CheckStatus(state.None.String())
}
