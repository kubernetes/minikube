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

var _ MiniTestTester = &KVMDockerLinuxAmd64IntegrationTester{}

// this runs the integration tests with kvm2 driver and a docker container runtime.
type KVMContainerdLinuxAmd64IntegrationTester struct {
}

// Run implements MiniTestTester.
func (k *KVMContainerdLinuxAmd64IntegrationTester) Run(runner MiniTestRunner) error {
	return kvmGeneralTester(runner, "./hack/prow/integration_kvm_containerd_linux_x86.sh")
}
