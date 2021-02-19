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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util/retry"

	"github.com/docker/machine/libmachine/state"
	"github.com/hashicorp/go-getter"
	pkgutil "k8s.io/minikube/pkg/util"
)

func installRelease(version string) (f *os.File, err error) {
	tf, err := ioutil.TempFile("", fmt.Sprintf("minikube-%s.*.exe", version))
	if err != nil {
		return tf, err
	}
	tf.Close()

	url := pkgutil.GetBinaryDownloadURL(version, runtime.GOOS)

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

// legacyStartArgs returns the arguments normally used for starting older versions of minikube
func legacyStartArgs() []string {
	return strings.Split(strings.Replace(*startArgs, "--driver", "--vm-driver", -1), " ")
}

// TestRunningBinaryUpgrade upgrades a running legacy cluster to head minikube
func TestRunningBinaryUpgrade(t *testing.T) {
	// not supported till v1.10, and passing new images to old releases isn't supported anyways
	if TestingKicBaseImage() {
		t.Skipf("Skipping, test does not make sense with --base-image")
	}

	MaybeParallel(t)
	profile := UniqueProfileName("running-upgrade")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(55))

	defer CleanupWithLogs(t, profile, cancel)

	// Should be a version from the last 6 months
	legacyVersion := "v1.6.2"
	if KicDriver() {
		// v1.8.0 would be selected, but: https://github.com/kubernetes/minikube/issues/8740
		legacyVersion = "v1.9.0"
	}

	tf, err := installRelease(legacyVersion)
	if err != nil {
		t.Fatalf("%s release installation failed: %v", legacyVersion, err)
	}
	defer os.Remove(tf.Name())

	args := append([]string{"start", "-p", profile, "--memory=2200"}, legacyStartArgs()...)
	rr := &RunResult{}
	r := func() error {
		c := exec.CommandContext(ctx, tf.Name(), args...)
		legacyEnv := []string{}
		// replace the global KUBECONFIG with a fresh kubeconfig
		// because for minikube<1.17.0 it can not read the new kubeconfigs that have extra "Extenions" block
		// see: https://github.com/kubernetes/minikube/issues/10210
		for _, e := range os.Environ() {
			if !strings.Contains(e, "KUBECONFIG") { // get all global envs except the Kubeconfig which is used by new versions of minikubes
				legacyEnv = append(legacyEnv, e)
			}
		}
		// using a fresh kubeconfig for this test
		legacyKubeConfig, err := ioutil.TempFile("", "legacy_kubeconfig")
		if err != nil {
			t.Fatalf("failed to create temp file for legacy kubeconfig %v", err)
		}

		defer os.Remove(legacyKubeConfig.Name()) // clean up
		legacyEnv = append(legacyEnv, fmt.Sprintf("KUBECONFIG=%s", legacyKubeConfig.Name()))
		c.Env = legacyEnv
		rr, err = Run(t, c)
		return err
	}

	// Retry up to two times, to allow flakiness for the legacy release
	if err := retry.Expo(r, 1*time.Second, Minutes(30), 2); err != nil {
		t.Fatalf("legacy %s start failed: %v", legacyVersion, err)
	}

	args = append([]string{"start", "-p", profile, "--memory=2200", "--alsologtostderr", "-v=1"}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("upgrade from %s to HEAD failed: %s: %v", legacyVersion, rr.Command(), err)
	}
}

// TestStoppedBinaryUpgrade starts a legacy minikube and stops it and then upgrades to head minikube
func TestStoppedBinaryUpgrade(t *testing.T) {
	// not supported till v1.10, and passing new images to old releases isn't supported anyways
	if TestingKicBaseImage() {
		t.Skipf("Skipping, test does not make sense with --base-image")
	}

	MaybeParallel(t)
	profile := UniqueProfileName("stopped-upgrade")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(55))

	defer CleanupWithLogs(t, profile, cancel)

	// Guarantee stopped upgrade compatibility from a release that is at least 1 year old
	// NOTE: <v1.4.0 does not automatically install a hyperkit/KVM driver
	legacyVersion := "v1.0.0"

	if KicDriver() {
		// first release with non-experimental KIC
		legacyVersion = "v1.8.0"
	}

	tf, err := installRelease(legacyVersion)
	if err != nil {
		t.Fatalf("%s release installation failed: %v", legacyVersion, err)
	}
	defer os.Remove(tf.Name())

	args := append([]string{"start", "-p", profile, "--memory=2200"}, legacyStartArgs()...)
	rr := &RunResult{}
	r := func() error {
		c := exec.CommandContext(ctx, tf.Name(), args...)
		legacyEnv := []string{}
		// replace the global KUBECONFIG with a fresh kubeconfig
		// because for minikube<1.17.0 it can not read the new kubeconfigs that have extra "Extenions" block
		// see: https://github.com/kubernetes/minikube/issues/10210
		for _, e := range os.Environ() {
			if !strings.Contains(e, "KUBECONFIG") { // get all global envs except the Kubeconfig which is used by new versions of minikubes
				legacyEnv = append(legacyEnv, e)
			}
		}
		// using a fresh kubeconfig for this test
		legacyKubeConfig, err := ioutil.TempFile("", "legacy_kubeconfig")
		if err != nil {
			t.Fatalf("failed to create temp file for legacy kubeconfig %v", err)
		}

		defer os.Remove(legacyKubeConfig.Name()) // clean up
		legacyEnv = append(legacyEnv, fmt.Sprintf("KUBECONFIG=%s", legacyKubeConfig.Name()))
		c.Env = legacyEnv
		rr, err = Run(t, c)
		return err
	}

	// Retry up to two times, to allow flakiness for the legacy release
	if err := retry.Expo(r, 1*time.Second, Minutes(30), 2); err != nil {
		t.Fatalf("legacy %s start failed: %v", legacyVersion, err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, tf.Name(), "-p", profile, "stop"))
	if err != nil {
		t.Errorf("failed to stop cluster: %s: %v", rr.Command(), err)
	}

	args = append([]string{"start", "-p", profile, "--memory=2200", "--alsologtostderr", "-v=1"}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Fatalf("upgrade from %s to HEAD failed: %s: %v", legacyVersion, rr.Command(), err)
	}

	t.Run("MinikubeLogs", func(t *testing.T) {
		rr, err = Run(t, exec.CommandContext(ctx, Target(), "logs"))
		if err != nil {
			t.Fatalf("`minikube logs` after upgrade to HEAD from %s failed: %v", legacyVersion, err)
		}
	})
}

// TestKubernetesUpgrade upgrades Kubernetes from oldest to newest
func TestKubernetesUpgrade(t *testing.T) {
	MaybeParallel(t)
	profile := UniqueProfileName("kubernetes-upgrade")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(55))

	defer CleanupWithLogs(t, profile, cancel)

	args := append([]string{"start", "-p", profile, "--memory=2200", fmt.Sprintf("--kubernetes-version=%s", constants.OldestKubernetesVersion), "--alsologtostderr", "-v=1"}, StartArgs()...)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to start minikube HEAD with oldest k8s version: %s: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "--format={{.Host}}"))
	if err != nil {
		t.Logf("status error: %v (may be ok)", err)
	}

	got := strings.TrimSpace(rr.Stdout.String())
	if got != state.Stopped.String() {
		t.Errorf("FAILED: status = %q; want = %q", got, state.Stopped.String())
	}

	args = append([]string{"start", "-p", profile, "--memory=2200", fmt.Sprintf("--kubernetes-version=%s", constants.NewestKubernetesVersion), "--alsologtostderr", "-v=1"}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to upgrade with newest k8s version. args: %s : %v", rr.Command(), err)
	}

	s, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "version", "--output=json"))
	if err != nil {
		t.Fatalf("error running kubectl: %v", err)
	}
	cv := struct {
		ServerVersion struct {
			GitVersion string `json:"gitVersion"`
		} `json:"serverVersion"`
	}{}
	err = json.Unmarshal(s.Stdout.Bytes(), &cv)

	if err != nil {
		t.Fatalf("error traversing json output: %v", err)
	}

	if cv.ServerVersion.GitVersion != constants.NewestKubernetesVersion {
		t.Fatalf("expected server version %s is not the same with latest version %s", cv.ServerVersion.GitVersion, constants.NewestKubernetesVersion)
	}

	t.Logf("Attempting to downgrade Kubernetes (should fail)")
	args = append([]string{"start", "-p", profile, "--memory=2200", fmt.Sprintf("--kubernetes-version=%s", constants.OldestKubernetesVersion)}, StartArgs()...)
	if rr, err := Run(t, exec.CommandContext(ctx, Target(), args...)); err == nil {
		t.Fatalf("downgrading Kubernetes should not be allowed. expected to see error but got %v for %q", err, rr.Command())
	}

	t.Logf("Attempting restart after unsuccessful downgrade")
	args = append([]string{"start", "-p", profile, "--memory=2200", fmt.Sprintf("--kubernetes-version=%s", constants.NewestKubernetesVersion), "--alsologtostderr", "-v=1"}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("start after failed upgrade: %s: %v", rr.Command(), err)
	}
}

// TestMissingContainerUpgrade tests a Docker upgrade where the underlying container is missing
func TestMissingContainerUpgrade(t *testing.T) {
	if !DockerDriver() {
		t.Skipf("This test is only for Docker")
	}

	// not supported till v1.10, and passing new images to old releases isn't supported anyways
	if TestingKicBaseImage() {
		t.Skipf("Skipping, test does not make sense with --base-image")
	}

	MaybeParallel(t)
	profile := UniqueProfileName("missing-upgrade")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(55))

	defer CleanupWithLogs(t, profile, cancel)

	legacyVersion := "v1.9.1"

	tf, err := installRelease(legacyVersion)
	if err != nil {
		t.Fatalf("%s release installation failed: %v", legacyVersion, err)
	}
	defer os.Remove(tf.Name())

	args := append([]string{"start", "-p", profile, "--memory=2200"}, StartArgs()...)
	rr := &RunResult{}
	r := func() error {
		rr, err = Run(t, exec.CommandContext(ctx, tf.Name(), args...))
		return err
	}

	// Retry up to two times, to allow flakiness for the previous release
	if err := retry.Expo(r, 1*time.Second, Minutes(30), 2); err != nil {
		t.Fatalf("release start failed: %v", err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, "docker", "stop", profile))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	rr, err = Run(t, exec.CommandContext(ctx, "docker", "rm", profile))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	args = append([]string{"start", "-p", profile, "--memory=2200", "--alsologtostderr", "-v=1"}, StartArgs()...)
	rr, err = Run(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed missing container upgrade from %s. args: %s : %v", legacyVersion, rr.Command(), err)
	}
}
