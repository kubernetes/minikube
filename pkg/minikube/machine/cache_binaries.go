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

package machine

import (
	"crypto"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/golang/glog"
	"github.com/jimmidyson/go-download"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
)

// CacheBinariesForBootstrapper will cache binaries for a bootstrapper
func CacheBinariesForBootstrapper(version string, clusterBootstrapper string) error {
	binaries := bootstrapper.GetCachedBinaryList(clusterBootstrapper)

	var g errgroup.Group
	for _, bin := range binaries {
		bin := bin
		g.Go(func() error {
			if _, err := CacheBinary(bin, version, "linux", runtime.GOARCH); err != nil {
				return errors.Wrapf(err, "caching image %s", bin)
			}
			return nil
		})
	}
	return g.Wait()
}

// CacheBinary will cache a binary on the host
func CacheBinary(binary, version, osName, archName string) (string, error) {
	glog.Infof("CacheBinary start: %s", binary)
	defer glog.Infof("CacheBinary end: %s", binary)
	dest := path.Join(constants.MakeMiniPath("cache", version), binary)
	_, err := os.Stat(dest)
	if err == nil {
		glog.Infof("Found local cache: %s", dest)
		return dest, nil
	}
	if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "stat")
	}
	url := constants.GetKubernetesReleaseURL(binary, version, osName, archName)
	console.OutStyle("file-download", "Downloading %s", url)

	options := download.FileOptions{Mkdirs: download.MkdirAll}
	options.Checksum = constants.GetKubernetesReleaseURLSHA1(binary, version, osName, archName)
	options.ChecksumHash = crypto.SHA1
	if err := download.ToFile(url, dest, options); err != nil {
		return "", errors.Wrapf(err, "Error downloading %s %s", binary, version)
	}
	if osName == runtime.GOOS && archName == runtime.GOARCH {
		if err = os.Chmod(dest, 0755); err != nil {
			return "", errors.Wrapf(err, "chmod +x %s", dest)
		}
	}
	return dest, nil
}

// CopyBinary copies previously cached binaries into the path
func CopyBinary(cmd bootstrapper.CommandRunner, src string, dest string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return errors.Wrap(err, "Stat")
	}
	dsize, err := cmd.FileSize(dest)
	if err == nil && dsize == fi.Size() {
		glog.Infof("%s already exists and is %d bytes", dest, dsize)
		return nil
	}

	f, err := assets.NewFileAsset(src, filepath.Dir(dest), filepath.Base(dest), "0755")
	if err != nil {
		return errors.Wrap(err, "new file asset")
	}
	if err := cmd.Copy(f); err != nil {
		return errors.Wrapf(err, "copy")
	}
	return nil
}
