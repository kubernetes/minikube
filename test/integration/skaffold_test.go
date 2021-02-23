// +build integration

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

package integration

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/otiai10/copy"
	"k8s.io/minikube/pkg/util/retry"
)

func TestSkaffold(t *testing.T) {
	if NativeDriver() {
		t.Skip("native driver doesn't support `minikube docker-env`; skaffold depends on this command")
	}
	if cr := ContainerRuntime(); cr != "docker" {
		t.Skipf("skaffold requires docker-env, currently testing %s container runtime", cr)
	}

	profile := UniqueProfileName("skaffold")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
	defer CleanupWithLogs(t, profile, cancel)

	// install latest skaffold release
	tf, err := installSkaffold()
	if err != nil {
		t.Fatalf("skaffold release installation failed: %v", err)
	}
	defer os.Remove(tf.Name())

	rr, err := Run(t, exec.CommandContext(ctx, tf.Name(), "version"))
	if err != nil {
		t.Fatalf("error running skaffold version: %v\n%s", err, rr.Output())
	}
	t.Logf("skaffold version: %s", rr.Stdout.Bytes())

	args := append([]string{"start", "-p", profile, "--memory=2600"}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("starting minikube: %v\n%s", err, rr.Output())
	}

	// make sure minikube binary is in path so that skaffold can access it
	abs, err := filepath.Abs(Target())
	if err != nil {
		t.Fatalf("unable to determine abs path: %v", err)
	}

	if filepath.Base(Target()) != "minikube" {
		new := filepath.Join(filepath.Dir(abs), "minikube")
		t.Logf("copying %s to %s", Target(), new)
		if err := copy.Copy(Target(), new); err != nil {
			t.Fatalf("error copying to minikube")
		}
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Dir(abs), os.Getenv("PATH")))

	// make sure 'docker' and 'minikube' are now in PATH
	for _, binary := range []string{"minikube", "docker"} {
		_, err := exec.LookPath(binary)
		if err != nil {
			t.Fatalf("%q is not in path", binary)
		}
	}

	defer func() {
		os.Setenv("PATH", oldPath)
	}()

	// make sure "skaffold run" exits without failure
	cmd := exec.CommandContext(ctx, tf.Name(), "run", "--minikube-profile", profile, "--kube-context", profile, "--status-check=true", "--port-forward=false", "--interactive=false")
	cmd.Dir = "testdata/skaffold"
	rr, err = Run(t, cmd)
	if err != nil {
		t.Fatalf("error running skaffold: %v\n%s", err, rr.Output())
	}

	// make sure expected deployment is running
	if _, err := PodWait(ctx, t, profile, "default", "app=leeroy-app", Minutes(1)); err != nil {
		t.Fatalf("failed waiting for pod leeroy-app: %v", err)
	}
	if _, err := PodWait(ctx, t, profile, "default", "app=leeroy-web", Minutes(1)); err != nil {
		t.Fatalf("failed waiting for pod leeroy-web: %v", err)
	}
}

// installSkaffold installs the latest release of skaffold
func installSkaffold() (f *os.File, err error) {
	tf, err := ioutil.TempFile("", "skaffold.exe")
	if err != nil {
		return tf, err
	}
	tf.Close()

	url := "https://storage.googleapis.com/skaffold/releases/latest/skaffold-%s-%s"
	url = fmt.Sprintf(url, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		url += ".exe"
	}

	if err := retry.Expo(func() error { return getter.GetFile(tf.Name(), url) }, 3*time.Second, Minutes(3)); err != nil {
		return tf, err
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(tf.Name(), 0700); err != nil {
			return tf, err
		}
	}
	return tf, nil
}
