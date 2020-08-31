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
	"path"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
)

// CacheBinariesForBootstrapper will cache binaries for a bootstrapper
func CacheBinariesForBootstrapper(version string, clusterBootstrapper string) error {
	binaries := bootstrapper.GetCachedBinaryList(clusterBootstrapper)

	// If --download-only, skip KubernetesReleaseBinaries
	if viper.GetBool("download-only") {
		var cBinaries []string
		for _, bin := range binaries {
			excluded := false
			for _, excludedBin := range constants.KubernetesReleaseBinaries {
				if bin == excludedBin {
					excluded = true
				}
			}
			if !excluded {
				cBinaries = append(cBinaries, bin)
			}
		}
		binaries = cBinaries
	}

	var g errgroup.Group
	for _, bin := range binaries {
		bin := bin // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			if _, err := download.Binary(bin, version, "linux", runtime.GOARCH); err != nil {
				return errors.Wrapf(err, "caching binary %s", bin)
			}
			return nil
		})
	}
	return g.Wait()
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
