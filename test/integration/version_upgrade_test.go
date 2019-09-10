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
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util/retry"

	"github.com/hashicorp/go-getter"
	pkgutil "k8s.io/minikube/pkg/util"
)

// TestVersionUpgrade downloads latest version of minikube and runs with
// the odlest supported k8s version and then runs the current head minikube
// and it tries to upgrade from the older supported k8s to news supported k8s
func TestVersionUpgrade(t *testing.T) {
	profile := Profile("vupgrade")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	MaybeParallel(t)

	defer CleanupWithLogs(t, profile, cancel)

	t.Logf("Downloading the latest release ...")
	start := time.Now()
	rpath, err := downloadLatestRelease()
	if err != nil {
		t.Fatalf("download minikube: %v", err)
	}
	t.Logf("Download completed within %s", time.Since(start))
	defer os.Remove(rpath)

	t.Logf("Starting last release with oldest Kubernetes version")
	rr, err := Run(ctx, t, rpath, "start", "-p", profile, fmt.Sprintf("--kubernetes-version=%s", constants.OldestKubernetesVersion))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	rr, err = Run(ctx, t, rpath, "stop", "-p", profile)
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Args, err)
	}

	rr, err = Run(ctx, t, rpath, "-p", profile, "status", "--format={{.Host}}")
	if err != nil {
		t.Logf("status error: %v (may be ok)", err)
	}
	got := strings.TrimSpace(rr.Stdout.String())
	if got != state.Stopped.String() {
		t.Errorf("status = %q; want = %q", got, state.Stopped.String())
	}

	t.Logf("Restarting cluster with %s and newest possible Kubernetes", Target())
	rr, err = Run(ctx, t, Target(), "start", "-p", profile, fmt.Sprintf("--kubernetes-version=%s", constants.NewestKubernetesVersion))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Args, err)
	}
}

func downloadLatestRelease() (string, error) {
	tf, err := ioutil.TempFile("", "minikube-release.*.exe")
	if err != nil {
		return tf.Name(), err
	}

	url := pkgutil.GetBinaryDownloadURL("latest", runtime.GOOS)
	download := func() error {
		return getter.GetFile(tf.Name(), url)
	}

	if err := retry.Expo(download, 3*time.Second, 3*time.Minute); err != nil {
		return tf.Name(), errors.Wrap(err, "Failed to get latest release binary")
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tf.Name(), 0700); err != nil {
			return tf.Name(), err
		}
	}
	return tf.Name(), nil
}
