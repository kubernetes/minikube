/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package kubeadm

import (
	"errors"
	"os/exec"
	"path"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
)

var errSyntheticInitialization = errors.New("synthetic initialization boundary")

type retainedControlPlaneRunner struct {
	*command.FakeCommandRunner
	t           *testing.T
	existing    bool
	retained    bool
	denied      bool
	initialized bool
}

// RunCmd admits metadata probes and rejects reset or initialization of retained state.
func (r *retainedControlPlaneRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	result := &command.RunResult{Args: cmd.Args}
	joined := strings.Join(cmd.Args, " ")
	if strings.Contains(joined, "kubeadm reset") {
		r.t.Fatal("restart failure reached destructive kubeadm reset")
	}
	if len(cmd.Args) > 1 && cmd.Args[1] == "ls" {
		if r.existing {
			return result, nil
		}
		result.ExitCode = 1
		return result, errors.New("synthetic missing configuration")
	}
	if len(cmd.Args) == 4 && cmd.Args[1] == "test" {
		if cmd.Args[2] == "-s" && cmd.Args[3] == path.Join(bsutil.EtcdDataDir(), "member/snap/db") {
			if r.denied {
				result.ExitCode = 255
				return result, errors.New("synthetic transport denial")
			}
			if r.retained {
				return result, nil
			}
		}
		result.ExitCode = 1
		return result, errors.New("synthetic absent file")
	}
	if len(cmd.Args) > 1 && cmd.Args[1] == "cp" {
		r.initialized = true
		return result, errSyntheticInitialization
	}
	return result, nil
}

// TestRetainedControlPlaneRestart holds reset when managed primary restart fails.
func TestRetainedControlPlaneRestart(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "enabled", false: "disabled"}[enabled], func(t *testing.T) {
			runner := &retainedControlPlaneRunner{FakeCommandRunner: command.NewFakeCommandRunner(), t: t, existing: true}
			bootstrap := &Bootstrapper{c: runner}
			cfg := config.ClusterConfig{Driver: "vfkit", ClusterAutoscaler: &config.ClusterAutoscalerConfig{}, Addons: map[string]bool{"cluster-autoscaler": enabled}, KubernetesConfig: config.KubernetesConfig{KubernetesVersion: "v1.37.2", ContainerRuntime: "containerd"}}
			// An incomplete primary-node selection forces a real restart error before network I/O.
			err := bootstrap.StartCluster(cfg, nil)
			if err == nil || !strings.Contains(err.Error(), "preserve") || runner.initialized {
				t.Fatalf("got error=%v initialized=%v; require preserved restart failure", err, runner.initialized)
			}
		})
	}
}

// TestRetainedControlPlaneConfiguration distinguishes data retention from fresh bootstrap.
func TestRetainedControlPlaneConfiguration(t *testing.T) {
	for _, scenario := range []string{"retained", "denied", "fresh"} {
		t.Run(scenario, func(t *testing.T) {
			runner := &retainedControlPlaneRunner{FakeCommandRunner: command.NewFakeCommandRunner(), t: t, retained: scenario == "retained", denied: scenario == "denied"}
			bootstrap := &Bootstrapper{c: runner}
			cfg := config.ClusterConfig{Driver: "vfkit", ClusterAutoscaler: &config.ClusterAutoscalerConfig{}, KubernetesConfig: config.KubernetesConfig{KubernetesVersion: "v1.37.2", ContainerRuntime: "containerd"}}
			err := bootstrap.StartCluster(cfg, nil)
			if scenario == "fresh" {
				if !errors.Is(err, errSyntheticInitialization) || !runner.initialized {
					t.Fatalf("fresh bootstrap was held: %v", err)
				}
			} else if err == nil || runner.initialized {
				t.Fatalf("got error=%v initialized=%v; require held retained state", err, runner.initialized)
			}
		})
	}
}
