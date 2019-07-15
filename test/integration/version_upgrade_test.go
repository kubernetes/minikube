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
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/state"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
	pkgutil "k8s.io/minikube/pkg/util"
)

func downloadMinikubeBinary(version string) (*os.File, error) {
	// Grab latest release binary
	url := pkgutil.GetBinaryDownloadURL(version, runtime.GOOS)
	resp, err := retryablehttp.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get latest release binary")
	}
	defer resp.Body.Close()

	tf, err := ioutil.TempFile("", "minikube")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create binary file")
	}
	_, err = io.Copy(tf, resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to populate temp file")
	}
	if err := tf.Close(); err != nil {
		return nil, errors.Wrap(err, "Failed to close temp file")
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tf.Name(), 0700); err != nil {
			return nil, err
			// t.Fatal(errors.Wrap(err, "Failed to make binary executable."))
		}
	}
	return tf, err
}

// TestVersionUpgrade downloads latest version of minikube and runs with
// the odlest supported k8s version and then runs the current head minikube
// and it tries to upgrade from the older supported k8s to news supported k8s
func TestVersionUpgrade(t *testing.T) {
	currentRunner := NewMinikubeRunner(t)
	currentRunner.RunCommand("delete", true)
	currentRunner.CheckStatus(state.None.String())
	tf, err := downloadMinikubeBinary("latest")
	if err != nil || tf == nil {
		t.Fatal(errors.Wrap(err, "Failed to download minikube binary."))
	}
	defer os.Remove(tf.Name())

	releaseRunner := NewMinikubeRunner(t)
	releaseRunner.BinaryPath = tf.Name()
	// For full coverage: also test upgrading from oldest to newest supported k8s release
	releaseRunner.Start(fmt.Sprintf("--kubernetes-version=%s", constants.OldestKubernetesVersion))
	releaseRunner.CheckStatus(state.Running.String())
	releaseRunner.RunCommand("stop", true)
	releaseRunner.CheckStatus(state.Stopped.String())

	// Trim the leading "v" prefix to assert that we handle it properly.
	currentRunner.Start(fmt.Sprintf("--kubernetes-version=%s", strings.TrimPrefix(constants.NewestKubernetesVersion, "v")))
	currentRunner.CheckStatus(state.Running.String())
	currentRunner.RunCommand("delete", true)
	currentRunner.CheckStatus(state.None.String())
}
