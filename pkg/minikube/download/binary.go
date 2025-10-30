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

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
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
	// Prevent Kubernetes binary downloads in --no-kubernetes mode
	if viper.GetBool("no-kubernetes") {
		klog.Infof("Skipping Kubernetes binary download due to --no-kubernetes flag")
		return "", nil
	}
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

	if osName == runtime.GOOS && archName == runtime.GOARCH {
		if err = os.Chmod(targetFilepath, 0755); err != nil {
			return "", errors.Wrapf(err, "chmod +x %s", targetFilepath)
		}
	}
	return targetFilepath, nil
}
