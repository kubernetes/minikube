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
	"os/exec"
	"testing"
	"time"

	commonutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func testClusterEnv(t *testing.T) {
	t.Parallel()

	minikubeRunner := util.MinikubeRunner{
		Args:       *args,
		BinaryPath: *binaryPath,
		T:          t}

	dockerEnvVars := minikubeRunner.RunCommand("docker-env", true)
	if err := minikubeRunner.SetEnvFromEnvCmdOutput(dockerEnvVars); err != nil {
		t.Fatalf("Error parsing output: %s", err)
	}
	path, err := exec.LookPath("docker")

	var output []byte
	dockerPs := func() error {
		cmd := exec.Command(path, "ps")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return &commonutil.RetriableError{Err: err}
		}
		return nil
	}
	if err := commonutil.RetryAfter(5, dockerPs, 3*time.Second); err != nil {
		t.Fatalf("Error running command: %s. Error: %s Output: %s", "docker ps", err, output)
	}
}
