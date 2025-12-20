/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package cruntime

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/localpath"
)

type MockRunner struct {
	*FakeRunner
	Output string
}

func (f *MockRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	f.FakeRunner.cmds = append(f.FakeRunner.cmds, cmd.Args...)
	args := cmd.Args
	if len(args) > 0 {
		if args[0] == "sudo" {
			args = args[1:]
		}
		// Handle mktemp
		if args[0] == "mktemp" && len(args) > 1 && args[1] == "-d" {
			return &command.RunResult{Stdout: *bytes.NewBufferString("/tmp/backup")}, nil
		}

		// Handle crictl images
		if len(args) > 0 && (strings.HasSuffix(args[0], "crictl") || args[0] == "crictl") {
			for _, arg := range args {
				if arg == "images" {
					if f.Output != "" {
						return &command.RunResult{Stdout: *bytes.NewBufferString(f.Output)}, nil
					}
					return &command.RunResult{Stdout: *bytes.NewBufferString(`{ "images": [] }`)}, nil
				}
			}
		}
	}
	return f.FakeRunner.RunCmd(cmd)
}

func TestCRIOPreload(t *testing.T) {
	viper.Set("preload", true)
	tempDir := t.TempDir()
	t.Setenv(localpath.MinikubeHome, tempDir)

	k8sVersion := semver.MustParse("1.25.0")
	cRuntime := "crio"

	tarballName := download.TarballName(k8sVersion.String(), cRuntime)
	tarballDir := localpath.MakeMiniPath("cache", "preloaded-tarball")
	if err := os.MkdirAll(tarballDir, 0755); err != nil {
		t.Fatalf("Failed to create tarball dir: %v", err)
	}
	tarballPath := filepath.Join(tarballDir, tarballName)
	if err := os.WriteFile(tarballPath, []byte("dummy-content"), 0644); err != nil {
		t.Fatalf("Failed to create dummy tarball: %v", err)
	}

	cc := config.ClusterConfig{
		KubernetesConfig: config.KubernetesConfig{
			KubernetesVersion: k8sVersion.String(),
			ContainerRuntime:  cRuntime,
			ImageRepository:   "registry.k8s.io",
		},
		Driver: "docker",
	}

	t.Run("ImagesExist", func(t *testing.T) {
		fake := NewFakeRunner(t)
		runner := &MockRunner{FakeRunner: fake}
		runner.Output = `{
			"images": [
				{
					"id": "sha256:1234567890",
					"repoTags": ["library/hello-world:latest"],
					"size": "123"
				}
			]
		}`

		r := &CRIO{
			Runner:            runner,
			KubernetesVersion: k8sVersion,
		}

		err := r.Preload(cc)
		if err != nil {
			t.Fatalf("Preload failed: %v", err)
		}

		// Verify sequence
		backupCalled := false
		saveCalled := false
		tarCalled := false
		loadCalled := false
		rmCalled := false

		for _, cmd := range fake.cmds {
			if cmd == "mktemp" {
				backupCalled = true
			}
			if cmd == "save" {
				saveCalled = true
			}
			if cmd == "tar" {
				tarCalled = true
			}
			if cmd == "load" {
				loadCalled = true
			}
			if cmd == "rm" {
				rmCalled = true
			}
		}

		if !backupCalled {
			t.Error("mktemp -d not called")
		}
		if !saveCalled {
			t.Error("podman save not called")
		}
		if !tarCalled {
			t.Error("tar not called")
		}
		if !loadCalled {
			t.Error("podman load not called")
		}
		if !rmCalled {
			t.Error("rm not called")
		}
	})

	t.Run("NoImages", func(t *testing.T) {
		fake := NewFakeRunner(t)
		runner := &MockRunner{FakeRunner: fake}
		runner.Output = `{ "images": [] }`

		r := &CRIO{
			Runner:            runner,
			KubernetesVersion: k8sVersion,
		}

		err := r.Preload(cc)
		if err != nil {
			t.Fatalf("Preload failed: %v", err)
		}

		// Verify only tar called
		saveCalled := false
		tarCalled := false
		for _, cmd := range fake.cmds {
			if cmd == "save" {
				saveCalled = true
			}
			if cmd == "tar" {
				tarCalled = true
			}
		}
		if saveCalled {
			t.Error("podman save called unexpectedly")
		}
		if !tarCalled {
			t.Error("tar not called")
		}
	})
}
