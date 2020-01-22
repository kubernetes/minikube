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
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/golang/glog"
	"github.com/jimmidyson/go-download"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
)

// CacheBinariesForBootstrapper will cache binaries for a bootstrapper
func CacheBinariesForBootstrapper(version string, clusterBootstrapper string) error {
	binaries := bootstrapper.GetCachedBinaryList(clusterBootstrapper)

	var g errgroup.Group
	for _, bin := range binaries {
		bin := bin // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			if _, err := CacheBinary(bin, version, "linux", runtime.GOARCH); err != nil {
				return errors.Wrapf(err, "caching binary %s", bin)
			}
			return nil
		})
	}
	return g.Wait()
}

// KubernetesReleaseURL gets the location of a kubernetes client
func KubernetesReleaseURL(binaryName, version, osName, archName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/%s", version, osName, archName, binaryName)
}

// KubernetesReleaseURLSHA1 gets the location of a kubernetes client checksum
func KubernetesReleaseURLSHA1(binaryName, version, osName, archName string) string {
	return fmt.Sprintf("%s.sha1", KubernetesReleaseURL(binaryName, version, osName, archName))
}

// CacheBinary will cache a binary on the host
func CacheBinary(binary, version, osName, archName string) (string, error) {
	targetDir := localpath.MakeMiniPath("cache", version)
	targetFilepath := path.Join(targetDir, binary)

	url := KubernetesReleaseURL(binary, version, osName, archName)

	_, err := os.Stat(targetFilepath)
	// If it exists, do no verification and continue
	if err == nil {
		glog.Infof("Not caching binary, using %s", url)
		return targetFilepath, nil
	}
	if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "stat %s version %s at %s", binary, version, targetDir)
	}

	if err = os.MkdirAll(targetDir, 0777); err != nil {
		return "", errors.Wrapf(err, "mkdir %s", targetDir)
	}

	options := download.FileOptions{
		Mkdirs: download.MkdirAll,
	}

	options.Checksum = KubernetesReleaseURLSHA1(binary, version, osName, archName)
	options.ChecksumHash = crypto.SHA1

	out.T(out.FileDownload, "Downloading {{.name}} {{.version}}", out.V{"name": binary, "version": version})
	if err := download.ToFile(url, targetFilepath, options); err != nil {
		return "", errors.Wrapf(err, "Error downloading %s %s", binary, version)
	}
	if osName == runtime.GOOS && archName == runtime.GOARCH {
		if err = os.Chmod(targetFilepath, 0755); err != nil {
			return "", errors.Wrapf(err, "chmod +x %s", targetFilepath)
		}
	}
	return targetFilepath, nil
}

// CopyBinary copies a locally cached binary to the guest VM
func CopyBinary(cr command.Runner, src string, dest string) error {
	f, err := assets.NewFileAsset(src, path.Dir(dest), path.Base(dest), "0755")
	if err != nil {
		return errors.Wrap(err, "new file asset")
	}
	if err := cr.Copy(f); err != nil {
		return errors.Wrapf(err, "copy")
	}
	return nil
}
