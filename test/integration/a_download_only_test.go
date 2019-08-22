// +build integration

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

// a_download_only_test.go filename starts with a, for the purpose that it runs before all parallel tests and downloads the images and caches them.
package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/retry"
)

// Note this test runs before all because filename is alphabetically first
// is used to cache images and binaries used by other parallel tests to avoid redownloading.
// TestDownloadOnly tests the --download-only option
func TestDownloadOnly(t *testing.T) {
	p := profileName(t)
	mk := NewMinikubeRunner(t, p)
	if !isTestNoneDriver(t) { // none driver doesnt need to be deleted
		defer mk.TearDown(t)
	}
	t.Run("group", func(t *testing.T) {
		t.Run("CacheOldestNewest", func(t *testing.T) {
			if isTestNoneDriver(t) { // don't cache images
				t.Skip("skipping test for none driver as it doesn't cache images")
			}

			minHome := constants.GetMinipath()
			for _, v := range []string{constants.OldestKubernetesVersion, constants.NewestKubernetesVersion} {
				mk.MustStart("--download-only", fmt.Sprintf("--kubernetes-version=%s", v))
				// checking if cached images are downloaded for example (kube-apiserver_v1.15.2, kube-scheduler_v1.15.2, ...)
				_, imgs := constants.GetKubeadmCachedImages("", v)
				for _, img := range imgs {
					img = strings.Replace(img, ":", "_", 1) // for example kube-scheduler:v1.15.2 --> kube-scheduler_v1.15.2
					fp := filepath.Join(minHome, "cache", "images", img)
					_, err := os.Stat(fp)
					if err != nil {
						t.Errorf("expected image file exist at %q but got error: %v", fp, err)
					}
				}

				// checking binaries downloaded (kubelet,kubeadm)
				for _, bin := range constants.GetKubeadmCachedBinaries() {
					fp := filepath.Join(minHome, "cache", v, bin)
					_, err := os.Stat(fp)
					if err != nil {
						t.Errorf("expected the file for binary exist at %q but got error %v", fp, err)
					}
				}
			}
		})
	})

	// this downloads the latest published binary from where we publish the minikube binary
	t.Run("DownloadLatestRelease", func(t *testing.T) {
		dest := filepath.Join(*testdataDir, fmt.Sprintf("minikube-%s-%s-latest-stable", runtime.GOOS, runtime.GOARCH))
		err := downloadMinikubeBinary(t, dest, "latest")
		if err != nil {
			t.Errorf("erorr downloading the latest minikube release %v", err)
		}
	})
}

// downloadMinikubeBinary downloads the minikube binary from github used by TestVersionUpgrade
// acts as a test setup for TestVersionUpgrade
func downloadMinikubeBinary(t *testing.T, dest string, version string) error {
	t.Helper()
	// Grab latest release binary
	url := pkgutil.GetBinaryDownloadURL(version, runtime.GOOS)
	download := func() error {
		return getter.GetFile(dest, url)
	}

	if err := retry.Expo(download, 3*time.Second, 3*time.Minute); err != nil {
		return errors.Wrap(err, "Failed to get latest release binary")
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dest, 0700); err != nil {
			return err
		}
	}
	return nil
}
