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

package browser

import (
	"os/exec"
	"runtime"

	"github.com/pkg/browser"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

// OpenURL opens a new browser window pointing to URL.
func OpenURL(url string) error {
	if runtime.GOOS == "linux" {
		_, err := exec.LookPath("xdg-open")
		if err != nil {
			out.Step(style.URL, url, false)
			return nil
		}
	}
	return browser.OpenURL(url)
}
