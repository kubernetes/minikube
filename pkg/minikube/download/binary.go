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

package download

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/util"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// DefaultKubeBinariesURL returns a URL to kube binaries
func DefaultKubeBinariesURL() string {
	return fmt.Sprintf("https://%s%s/release", releaseHost, releasePath)
}

// binaryWithChecksumURL gets the location of a Kubernetes binary
func binaryWithChecksumURL(binaryName, version, osName, archName, binaryURL string) (string, error) {
	if binaryURL == "" {
		binaryURL = DefaultKubeBinariesURL()
	}

	base := fmt.Sprintf("%s/%s/bin/%s/%s/%s", binaryURL, version, osName, archName, binaryName)
	v, err := semver.Make(version[1:])
	if err != nil {
		return "", err
	}

	if v.GTE(semver.MustParse("1.17.0")) {
		return fmt.Sprintf("%s?checksum=file:%s.sha256", base, base), nil
	}
	return fmt.Sprintf("%s?checksum=file:%s.sha1", base, base), nil
}

// Binary will download a binary onto the host
func Binary(binary, version, osName, archName, binaryURL string) (string, error) {
	targetDir := localpath.MakeMiniPath("cache", osName, archName, version)
	targetFilepath := path.Join(targetDir, binary)
	targetLock := targetFilepath + ".lock"

	url, err := binaryWithChecksumURL(binary, version, osName, archName, binaryURL)
	if err != nil {
		return "", err
	}

	releaser, err := lockDownload(targetLock)
	if releaser != nil {
		defer releaser.Release()
	}
	if err != nil {
		return "", err
	}

	if _, err := checkCache(targetFilepath); err == nil {
		klog.Infof("Not caching binary, using %s", url)
		return targetFilepath, nil
	}

	if err := download(url, targetFilepath); err != nil {
		return "", errors.Wrapf(err, "download failed: %s", url)
	}

	if osName == runtime.GOOS && archName == detect.EffectiveArch() {
		if err = os.Chmod(targetFilepath, 0755); err != nil {
			return "", errors.Wrapf(err, "chmod +x %s", targetFilepath)
		}
	}
	return targetFilepath, nil
}

// CrictlBinary download the crictl tar archive to the cache folder and untar it
func CrictlBinary(k8sversion string, crictlVersion string) (string, error) {
	// first we check whether crictl exists
	targetDir := localpath.MakeMiniPath("cache", "linux", runtime.GOARCH, k8sversion)
	targetPath := path.Join(targetDir, "crictl")
	if _, err := checkCache(targetPath); err == nil {
		klog.Infof("crictl found in cache: %s", targetPath)
		return targetPath, nil
	}
	v, err := util.ParseKubernetesVersion(k8sversion)
	if err != nil {
		return "", err
	}
	// if we don't know the exact patch number of crictl then use 0.
	// This definitely exists
	if crictlVersion == "" {
		crictlVersion = fmt.Sprintf("v%d.%d.%s", v.Major, v.Minor, crictlVersion)
	}
	url := fmt.Sprintf(
		"https://github.com/kubernetes-sigs/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz",
		crictlVersion, crictlVersion, runtime.GOARCH)
	if err := download(url, targetPath); err != nil {
		return "", errors.Wrapf(err, "download failed: %s", url)
	}

	return targetPath, nil
}
