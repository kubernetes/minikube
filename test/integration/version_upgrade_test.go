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

package integration

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func downloadMinikubeBinary(dest string, version string) error {
	// Grab latest release binary
	url := pkgutil.GetBinaryDownloadURL(version, runtime.GOOS)
	download := func() error {
		return getter.GetFile(dest, url)
	}

	if err := util.Retry2(download, 3*time.Second, 13); err != nil {
		return errors.Wrap(err, "Failed to get latest release binary")
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dest, 0700); err != nil {
			return err
		}
	}
	return nil
}

func fileExists(fname string) error {
	check := func() error {
		info, err := os.Stat(fname)
		if os.IsNotExist(err) {
			return err
		}
		if info.IsDir() {
			return fmt.Errorf("Error expect file got dir")
		}
		return nil
	}

	if err := util.Retry2(check, 3*time.Second, 13); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed check if file (%q) exists,", fname))
	}
	return nil
}

// TestVersionUpgrade downloads latest version of minikube and runs with
// the odlest supported k8s version and then runs the current head minikube
// and it tries to upgrade from the older supported k8s to news supported k8s
func TestVersionUpgrade(t *testing.T) {
	p := t.Name()
	err := fileExists("minikube_latest_binary")
	if err != nil {
		t.Fail()
	}

	mkCurrent := NewMinikubeRunner(t, p)
	if usingNoneDriver(mkCurrent) { // TODO (medyagh@): bring back once soled https://github.com/kubernetes/minikube/issues/4418
		t.Skip("skipping test as none driver does not support persistence")
	}
	mkCurrent.RunCommand("delete", true)
	mkCurrent.CheckStatus(state.None.String())

	fname := "minikube_latest_binary"
	defer os.Remove(fname)

	mkRelease := NewMinikubeRunner(t, p)
	mkRelease.BinaryPath = fname
	// For full coverage: also test upgrading from oldest to newest supported k8s release
	mkRelease.Start(fmt.Sprintf("--kubernetes-version=%s", constants.OldestKubernetesVersion))
	mkRelease.CheckStatus(state.Running.String())
	mkRelease.RunCommand("stop", true)
	mkRelease.CheckStatus(state.Stopped.String())

	opts := ""
	if usingNoneDriver(mkCurrent) { // to avoid https://github.com/kubernetes/minikube/issues/4418
		opts = "--apiserver-port=8444"
	}
	// Trim the leading "v" prefix to assert that we handle it properly.
	mkCurrent.Start(fmt.Sprintf("--kubernetes-version=%s", strings.TrimPrefix(constants.NewestKubernetesVersion, "v")), opts)
	mkCurrent.CheckStatus(state.Running.String())
	mkCurrent.RunCommand("delete", true)
	mkCurrent.CheckStatus(state.None.String())
}
