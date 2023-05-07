/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package drivers

import (
	"fmt"
	"os/exec"

	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/mcnutils"
)

// x7TODO:
// this is some slow logic... at least make it non-blocking..
// WaitForPrompt tries to run a command to the machine shell
// for 30 seconds before timing out
func WaitForPrompt(d Driver) error {
	if err := mcnutils.WaitFor(promptAvailFunc(d)); err != nil {
		return fmt.Errorf("Too many retries waiting for prompt to be available.  Last error: %s", err)
	}
	return nil
}

func promptAvailFunc(d Driver) func() bool {
	return func() bool {
		log.Debug("Getting to WaitForPrompt function...")
		if _, err := d.RunCmd(exec.Command("bash", "-c", "exit 0")); err != nil {
			log.Debugf("Error running 'exit 0' command : %s", err)
			return false
		}
		return true
	}
}
