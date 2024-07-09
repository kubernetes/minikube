/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"k8s.io/minikube/pkg/version"
)

func Licenses(dir string) error {
	resp, err := http.Get(fmt.Sprintf("https://storage.googleapis.com/minikube/releases/%s/licenses.tar.gz", version.GetVersion()))
	if err != nil {
		return fmt.Errorf("failed to download licenses: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download request did not return a 200, received: %d", resp.StatusCode)
	}
	f, err := os.CreateTemp("", "licenses")
	if err != nil {
		return fmt.Errorf("failed to create file in tmp dir: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("failed to copy: %v", err)
	}
	if err := exec.Command("tar", "-xvzf", f.Name(), "-C", dir).Run(); err != nil {
		return fmt.Errorf("failed to untar licenses: %v", err)
	}
	return nil
}
