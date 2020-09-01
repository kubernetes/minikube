/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/download"
)

type copyFailRunner struct {
	command.Runner
}

func (copyFailRunner) Copy(a assets.CopyableFile) error {
	return fmt.Errorf("test error during copy file")
}

func newFakeCommandRunnerCopyFail() command.Runner {
	return copyFailRunner{command.NewFakeCommandRunner()}
}

func TestCopyBinary(t *testing.T) {
	tc := []struct {
		lastUpdateCheckFilePath string
		src, dst, desc          string
		err                     bool
		runner                  command.Runner
	}{
		{
			desc:   "not existing src",
			dst:    "/tmp/testCopyBinary1",
			src:    "/tmp/testCopyBinary2",
			err:    true,
			runner: command.NewFakeCommandRunner(),
		},
		{
			desc:   "src /etc/hosts",
			dst:    "/tmp/testCopyBinary1",
			src:    "/etc/hosts",
			err:    false,
			runner: command.NewFakeCommandRunner(),
		},
		{
			desc:   "existing src, copy fail",
			dst:    "/etc/passwd",
			src:    "/etc/hosts",
			err:    true,
			runner: newFakeCommandRunnerCopyFail(),
		},
	}
	for _, test := range tc {
		t.Run(test.desc, func(t *testing.T) {
			err := CopyBinary(test.runner, test.src, test.dst)
			if err != nil && !test.err {
				t.Fatalf("Error %v expected but not occurred", err)
			}
			if err == nil && test.err {
				t.Fatal("Unexpected error")
			}
		})
	}
}

func TestCacheBinariesForBootstrapper(t *testing.T) {
	download.EnableMock(true)

	oldMinikubeHome := os.Getenv("MINIKUBE_HOME")
	defer os.Setenv("MINIKUBE_HOME", oldMinikubeHome)

	minikubeHome, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		t.Fatalf("error during creating tmp dir: %v", err)
	}

	defer func() { // clean up tempdir
		err := os.RemoveAll(minikubeHome)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", minikubeHome)
		}
	}()

	tc := []struct {
		version, clusterBootstrapper string
		minikubeHome                 string
		err                          bool
	}{
		{
			version:             "v1.16.0",
			clusterBootstrapper: bootstrapper.Kubeadm,
			err:                 false,
			minikubeHome:        minikubeHome,
		},
		{
			version:             "invalid version",
			clusterBootstrapper: bootstrapper.Kubeadm,
			err:                 true,
			minikubeHome:        minikubeHome,
		},
	}
	for _, test := range tc {
		t.Run(test.version, func(t *testing.T) {
			os.Setenv("MINIKUBE_HOME", test.minikubeHome)
			err := CacheBinariesForBootstrapper(test.version, test.clusterBootstrapper)
			if err != nil && !test.err {
				t.Fatalf("Got unexpected error %v", err)
			}
			if err == nil && test.err {
				t.Fatalf("Expected error but got %v", err)
			}
		})
	}
}
