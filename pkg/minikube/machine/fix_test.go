/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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
	"log"
	"os"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

func TestMaybeWarnAboutEvalEnv(t *testing.T) {
	orgShellEnv := os.Getenv("SHELL")
	defer os.Setenv("SHELL", orgShellEnv)

	// Capture stdout
	old := os.Stdout
	defer func() {
		os.Stdout = old
	}()

	var testCases = []struct {
		name             string
		activationMarker string
		expected         string
		shell            string
	}{
		{`docker`, constants.MinikubeActiveDockerdEnv, `eval $(minikube -p minikube docker-env)`, "bash"},
		{`podman`, constants.MinikubeActivePodmanEnv, `eval $(minikube -p minikube podman-env)`, "bash"},
		{`docker`, constants.MinikubeActiveDockerdEnv, `minikube -p minikube docker-env | Invoke-Expression`, "powershell"},
		{`podman`, constants.MinikubeActivePodmanEnv, `minikube -p minikube podman-env | Invoke-Expression`, "powershell"},
		{`docker`, constants.MinikubeActiveDockerdEnv, `minikube -p minikube docker-env | source`, "fish"},
		{`podman`, constants.MinikubeActivePodmanEnv, `minikube -p minikube podman-env | source`, "fish"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-%s", tc.shell, tc.name), func(t *testing.T) {
			os.Setenv("SHELL", tc.shell)
			tmpfile, err := ioutil.TempFile("", "machine_fix_test")
			if err != nil {
				log.Fatal(err)
			}

			defer os.Remove(tmpfile.Name())
			defer os.Setenv(tc.activationMarker, "")

			os.Stderr = tmpfile

			os.Setenv(tc.activationMarker, "1")

			cc := config.ClusterConfig{Name: "minikube"}
			maybeWarnAboutEvalEnv(tc.name, cc.Name)

			warningMsg, err := ioutil.ReadFile(tmpfile.Name())
			if err != nil {
				t.Fatalf("Unable to read file: %v", err)
			}
			os.Stdout = old
			msg := string(warningMsg)
			if !strings.Contains(msg, tc.expected) {
				t.Errorf("Expected that string: \"%s\" contains: \"%s\"", msg, tc.expected)
			}
		})
	}
}
