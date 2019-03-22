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
	"os/exec"
	"runtime"
	"testing"

	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/minikube/test/integration/util"
)

func TestVersionUpgrade(t *testing.T) {
	currentRunner := NewMinikubeRunner(t)
	currentRunner.RunCommand("delete", true)
	currentRunner.CheckStatus(state.None.String())

	isoName := "minikube-linux-amd64"
	if runtime.GOOS == "darwin" {
		isoName = "minikube-darwin-amd64"
	} else if runtime.GOOS == "windows" {
		isoName = "minikube-windows-amd64.exe"
	}
	// Grab latest release binary
	url := "https://storage.googleapis.com/minikube/releases/latest/" + isoName
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Failed to get latest release binary"))
	}
	defer resp.Body.Close()

	iso, err := ioutil.TempFile("", isoName)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Failed to create binary file"))
	}
	defer os.Remove(iso.Name())

	_, err = io.Copy(iso, resp.Body)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Failed to populate iso file"))
	}
	if err := iso.Close(); err != nil {
		t.Fatal(errors.Wrap(err, "Failed to close iso"))
	}

	if runtime.GOOS != "windows" {
		if err := exec.Command("chmod", "+x", iso.Name()).Run(); err != nil {
			t.Fatal(errors.Wrap(err, "Failed to make binary executable."))
		}
	}

	latestRunner := util.MinikubeRunner{
		Args:       currentRunner.Args,
		BinaryPath: iso.Name(),
		StartArgs:  currentRunner.StartArgs,
		MountArgs:  currentRunner.MountArgs,
		T:          t,
	}
	latestRunner.Start()
	latestRunner.CheckStatus(state.Running.String())
	latestRunner.RunCommand("stop", true)
	latestRunner.CheckStatus(state.Stopped.String())

	currentRunner.Start()
	currentRunner.CheckStatus(state.Running.String())
	currentRunner.RunCommand("delete", true)
	currentRunner.CheckStatus(state.None.String())
}
