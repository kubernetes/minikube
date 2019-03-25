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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	pkgutil "k8s.io/minikube/pkg/util"
	"k8s.io/minikube/test/integration/util"
)

func TestVersionUpgrade(t *testing.T) {
	currentRunner := NewMinikubeRunner(t)
	currentRunner.RunCommand("delete", true)
	currentRunner.CheckStatus(state.None.String())

	// Grab latest release binary
	url := pkgutil.GetBinaryDownloadURL("latest", runtime.GOOS)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Failed to get latest release binary"))
	}
	defer resp.Body.Close()

	tf, err := ioutil.TempFile("", "minikube")
	if err != nil {
		t.Fatal(errors.Wrap(err, "Failed to create binary file"))
	}
	defer os.Remove(tf.Name())

	_, err = io.Copy(tf, resp.Body)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Failed to populate temp file"))
	}
	if err := tf.Close(); err != nil {
		t.Fatal(errors.Wrap(err, "Failed to close temp file"))
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tf.Name(), 0700); err != nil {
			t.Fatal(errors.Wrap(err, "Failed to make binary executable."))
		}
	}

	releaseRunner := util.MinikubeRunner{
		Args:       currentRunner.Args,
		BinaryPath: tf.Name(),
		StartArgs:  currentRunner.StartArgs,
		MountArgs:  currentRunner.MountArgs,
		T:          t,
	}
	releaseRunner.Start()
	releaseRunner.CheckStatus(state.Running.String())
	releaseRunner.RunCommand("stop", true)
	releaseRunner.CheckStatus(state.Stopped.String())

	currentRunner.Start()
	currentRunner.CheckStatus(state.Running.String())
	currentRunner.RunCommand("delete", true)
	currentRunner.CheckStatus(state.None.String())
}
