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
	"os"

	"k8s.io/klog/v2"
)

var _ MiniTestTester = &KVMIntegrationTester{}

type KVMIntegrationTester struct {
}

// Run implements MiniTestTester.
func (k *KVMIntegrationTester) Run(runner MiniTestRunner) error {

	if up, err := runner.IsUp(); err != nil || !up {
		klog.Errorf("tester: deployed environment is not up: %v", err)
	}
	if err := runner.SyncToRemote(".", "~/minikube", []string{".cache", ".git"}); err != nil {
		klog.Errorf("failed to sync file in docker deployer: %v", err)
	}

	var testErr error
	// install docker and libvirtd first then run the test in a new shell
	if err := runner.Execute("cd minikube && ./hack/prow/linux_integration_kvm_pre.sh"); err != nil {
		klog.Errorf("failed to install docker in env: %v", err)
		return err
	}
	if testErr = runner.Execute("cd minikube && ./hack/prow/linux_integration_kvm.sh"); testErr != nil {
		klog.Errorf("failed to execute command in env: %v", testErr)
		// don't return here, we still want to collect the test reports
	}

	artifactLocation := os.Getenv("ARTIFACTS")
	klog.Infof("copying to %s", artifactLocation)

	if err := runner.SyncToHost("~/minikube/testout.txt", artifactLocation, nil); err != nil {
		klog.Errorf("failed to sync testout.txt from deployer: %v", err)
		return err
	}
	if err := runner.SyncToHost("~/minikube/test.json", artifactLocation, nil); err != nil {
		klog.Errorf("failed to sync test.json in from deployer: %v", err)
		return err
	}

	if err := runner.SyncToHost("~/minikube/junit-unit.xml", artifactLocation, nil); err != nil {
		klog.Errorf("failed to sync junit-unit.xml in from deployer: %v", err)
		return err
	}

	return testErr

}
