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
	"os"
	"os/exec"
	"testing"
	"time"

	"k8s.io/minikube/test/integration/util"
)

// Assert that docker-env subcommand outputs usable information for "docker ps"
func testClusterEnv(t *testing.T) {
	t.Parallel()

	r := NewMinikubeRunner(t)

	// Set a specific shell syntax so that we don't have to handle every possible user shell
	envOut := r.RunCommand("docker-env --shell=bash", true)
	vars := r.ParseEnvCmdOutput(envOut)
	if len(vars) == 0 {
		t.Fatalf("Failed to parse env vars:\n%s", envOut)
	}
	for k, v := range vars {
		t.Logf("Found: %s=%s", k, v)
		if err := os.Setenv(k, v); err != nil {
			t.Errorf("failed to set %s=%s: %v", k, v, err)
		}
	}

	path, err := exec.LookPath("docker")
	if err != nil {
		t.Fatalf("Unable to complete test: docker is not installed in PATH")
	}
	t.Logf("Using docker installed at %s", path)

	var output []byte
	dockerPs := func() error {
		cmd := exec.Command(path, "ps")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return err
		}
		return nil
	}
	if err := util.Retry(t, dockerPs, 3*time.Second, 5); err != nil {
		t.Fatalf("Error running command: %s. Error: %v Output: %s", "docker ps", err, output)
	}
}
