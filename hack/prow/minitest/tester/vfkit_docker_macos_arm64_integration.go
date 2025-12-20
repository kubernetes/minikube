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

package tester

import (
	"fmt"
	"k8s.io/klog/v2"
	"os"
)

var _ MiniTestTester = &VfkitDockerMacOSARM64IntegrationTester{}

// this runs the integration tests with vfkit driver and a docker container runtime.
type VfkitDockerMacOSARM64IntegrationTester struct {
}

// Run implements MiniTestTester.
func (k *VfkitDockerMacOSARM64IntegrationTester) Run(runner MiniTestRunner) error {
	if up, err := runner.IsUp(); err != nil || !up {
		klog.Errorf("tester: deployed environment is not up: %v", err)
	}

	defer func() {
		if err := runner.Execute("sudo rm -rf ~/minikube"); err != nil {
			klog.Errorf("failed to delete minikube in vfkit docker macos arm64 tester: %v", err)
		}
	}()

	if err := runner.SyncToRemote(".", "~/minikube", []string{".cache",".modcache"}); err != nil {
		klog.Errorf("failed to sync file in docker deployer: %v", err)
	}
	pr := os.Getenv("PULL_NUMBER")
	var testErr error

	if testErr = runner.Execute(fmt.Sprintf("cd minikube && PULL_NUMBER=\"%s\" %s", pr, "./hack/prow/integration_vfkit_docker_macos_arm64.sh")); testErr != nil {
		klog.Errorf("failed to execute command in env: %v", testErr)
		// don't return here, we still want to collect the test reports
	}

	// prow requires result file to be copied to $ARTIFACTS. All other files will not be persisted.
	if err := copyFileToArtifact(runner); err != nil {
		return err
	}
	return testErr
}
