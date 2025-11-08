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

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/version"
)

// licensesTarballURL returns the URL for the licenses tarball,
// with a fallback to Google Cloud Storage.
func licensesTarballURL() string {
	githubURL := fmt.Sprintf("https://github.com/kubernetes/minikube/releases/download/%s/licenses.tar.gz", version.GetVersion())
	gcsURL := fmt.Sprintf("https://storage.googleapis.com/minikube/releases/%s/licenses.tar.gz", version.GetVersion())

	// Try GitHub first. (check if the URL is reachable without downloading the file)
	resp, err := http.Head(githubURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		return githubURL
	}
	klog.Warningf("Failed to find licenses on GitHub (%s), falling back to Google Cloud Storage", err)

	// Fallback to Google Cloud Storage.
	return gcsURL
}

// Licenses downloads the licenses tarball and extracts its contents to the specified directory.
func Licenses(dir string) error {
	url := licensesTarballURL()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download licenses from %s: %v", url, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			klog.Warningf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download request to %s did not return a 200, received: %d", url, resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "licenses-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			klog.Warningf("Failed to remove temp file %s: %v", tempFile.Name(), err)
		}
	}()
	defer func() {
		if err := tempFile.Close(); err != nil {
			klog.Warningf("Failed to close temp file %s: %v", tempFile.Name(), err)
		}
	}()

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		return fmt.Errorf("failed to copy downloaded content from %s: %v", url, err)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	if err := exec.Command("tar", "-xvzf", tempFile.Name(), "-C", dir).Run(); err != nil {
		return fmt.Errorf("failed to untar licenses: %v", err)
	}

	return nil
}
