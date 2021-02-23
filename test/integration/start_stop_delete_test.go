// +build integration

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

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/machine/libmachine/state"
	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/constants"
)

func TestStartStop(t *testing.T) {
	MaybeParallel(t)

	t.Run("group", func(t *testing.T) {
		tests := []struct {
			name    string
			version string
			args    []string
		}{
			{"old-k8s-version", constants.OldestKubernetesVersion, []string{
				// default is the network created by libvirt, if we change the name minikube won't boot
				// because the given network doesn't exist
				"--kvm-network=default",
				"--kvm-qemu-uri=qemu:///system",
				"--disable-driver-mounts",
				"--keep-context=false",
			}},
			{"newest-cni", constants.NewestKubernetesVersion, []string{
				"--feature-gates",
				"ServerSideApply=true",
				"--network-plugin=cni",
				"--extra-config=kubelet.network-plugin=cni",
				"--extra-config=kubeadm.pod-network-cidr=192.168.111.111/16",
			}},
			{"default-k8s-different-port", constants.DefaultKubernetesVersion, []string{
				"--apiserver-port=8444",
			}},
			{"no-preload", constants.NewestKubernetesVersion, []string{
				"--preload=false",
			}},
			{"disable-driver-mounts", constants.DefaultKubernetesVersion, []string{
				"--disable-driver-mounts",
				"--extra-config=kubeadm.ignore-preflight-errors=SystemVerification",
			}},
			{"embed-certs", constants.DefaultKubernetesVersion, []string{
				"--embed-certs",
			}},
		}

		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				profile := UniqueProfileName(tc.name)
				ctx, cancel := context.WithTimeout(context.Background(), Minutes(30))
				defer Cleanup(t, profile, cancel)
				type validateStartStopFunc func(context.Context, *testing.T, string, string, string, []string)
				if !strings.Contains(tc.name, "docker") && NativeDriver() {
					t.Skipf("skipping %s - incompatible with native driver", t.Name())
				}
				if strings.Contains(tc.name, "disable-driver-mounts") && !VirtualboxDriver() {
					t.Skipf("skipping %s - only runs on virtualbox", t.Name())
				}

				waitFlag := "--wait=true"
				if strings.Contains(tc.name, "cni") { // wait=app_running is broken for CNI https://github.com/kubernetes/minikube/issues/7354
					waitFlag = "--wait=apiserver,system_pods,default_sa"
				}

				startArgs := append([]string{"start", "-p", profile, "--memory=2200", "--alsologtostderr", waitFlag}, tc.args...)
				startArgs = append(startArgs, StartArgs()...)
				startArgs = append(startArgs, fmt.Sprintf("--kubernetes-version=%s", tc.version))

				t.Run("serial", func(t *testing.T) {
					serialTests := []struct {
						name      string
						validator validateStartStopFunc
					}{
						{"FirstStart", validateFirstStart},
						{"DeployApp", validateDeploying},
						{"Stop", validateStop},
						{"EnableAddonAfterStop", validateEnableAddonAfterStop},
						{"SecondStart", validateSecondStart},
						{"UserAppExistsAfterStop", validateAppExistsAfterStop},
						{"AddonExistsAfterStop", validateAddonAfterStop},
						{"VerifyKubernetesImages", validateKubernetesImages},
						{"Pause", validatePauseAfterStart},
					}
					for _, stc := range serialTests {
						if ctx.Err() == context.DeadlineExceeded {
							t.Fatalf("Unable to run more tests (deadline exceeded)")
						}

						tcName := tc.name
						tcVersion := tc.version
						stc := stc

						t.Run(stc.name, func(t *testing.T) {
							stc.validator(ctx, t, profile, tcName, tcVersion, startArgs)
						})
					}

					if *cleanup {
						// Normally handled by cleanuprofile, but not fatal there
						rr, err := Run(t, exec.CommandContext(ctx, Target(), "delete", "-p", profile))
						if err != nil {
							t.Errorf("failed to clean up: args %q: %v", rr.Command(), err)
						}

						rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "config", "get-contexts", profile))
						if err != nil {
							t.Logf("config context error: %v (may be ok)", err)
						}
						if rr.ExitCode != 1 {
							t.Errorf("expected exit code 1, got %d. output: %s", rr.ExitCode, rr.Output())
						}
					}

				})

			})
		}
	})
}

func validateFirstStart(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	defer PostMortemLogs(t, profile)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Fatalf("failed starting minikube -first start-. args %q: %v", rr.Command(), err)
	}
}

func validateDeploying(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	defer PostMortemLogs(t, profile)
	if !strings.Contains(tcName, "cni") {
		testPodScheduling(ctx, t, profile)
	}
}

func validateStop(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	defer PostMortemLogs(t, profile)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "stop", "-p", profile, "--alsologtostderr", "-v=3"))
	if err != nil {
		t.Fatalf("failed stopping minikube - first stop-. args %q : %v", rr.Command(), err)
	}
}

func validateEnableAddonAfterStop(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	defer PostMortemLogs(t, profile)
	// The native driver never really stops
	if !NativeDriver() {
		got := Status(ctx, t, Target(), profile, "Host", profile)
		if got != state.Stopped.String() {
			t.Errorf("expected post-stop host status to be -%q- but got *%q*", state.Stopped, got)
		}
	}

	// Enable an addon to assert it comes up afterwards
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "addons", "enable", "dashboard", "-p", profile))
	if err != nil {
		t.Errorf("failed to enable an addon post-stop. args %q: %v", rr.Command(), err)
	}

}

func validateSecondStart(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	defer PostMortemLogs(t, profile)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		// Explicit fatal so that failures don't move directly to deletion
		t.Fatalf("failed to start minikube post-stop. args %q: %v", rr.Command(), err)
	}

	got := Status(ctx, t, Target(), profile, "Host", profile)
	if got != state.Running.String() {
		t.Errorf("expected host status after start-stop-start to be -%q- but got *%q*", state.Running, got)
	}

}

// validateAppExistsAfterStop verifies that a user's app will not vanish after a minikube stop
func validateAppExistsAfterStop(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	defer PostMortemLogs(t, profile)
	if strings.Contains(tcName, "cni") {
		t.Logf("WARNING: cni mode requires additional setup before pods can schedule :(")
	} else if _, err := PodWait(ctx, t, profile, "kubernetes-dashboard", "k8s-app=kubernetes-dashboard", Minutes(9)); err != nil {
		t.Errorf("failed waiting for 'addon dashboard' pod post-stop-start: %v", err)
	}

}

// validateAddonAfterStop validates that an addon which was enabled when minikube is stopped will be enabled and working..
func validateAddonAfterStop(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	defer PostMortemLogs(t, profile)
	if strings.Contains(tcName, "cni") {
		t.Logf("WARNING: cni mode requires additional setup before pods can schedule :(")
	} else if _, err := PodWait(ctx, t, profile, "kubernetes-dashboard", "k8s-app=kubernetes-dashboard", Minutes(9)); err != nil {
		t.Errorf("failed waiting for 'addon dashboard' pod post-stop-start: %v", err)
	}
}

func validateKubernetesImages(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	if !NativeDriver() {
		testPulledImages(ctx, t, profile, tcVersion)
	}
}

func validatePauseAfterStart(ctx context.Context, t *testing.T, profile string, tcName string, tcVersion string, startArgs []string) {
	defer PostMortemLogs(t, profile)
	testPause(ctx, t, profile)
}

// testPodScheduling asserts that this configuration can schedule new pods
func testPodScheduling(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	t.Helper()

	// schedule a pod to assert persistence
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "-f", filepath.Join(*testdataDir, "busybox.yaml")))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	// 8 minutes, because 4 is not enough for images to pull in all cases.
	names, err := PodWait(ctx, t, profile, "default", "integration-test=busybox", Minutes(8))
	if err != nil {
		t.Fatalf("wait: %v", err)
	}

	// Use this pod to confirm that the runtime resource limits are sane
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", names[0], "--", "/bin/sh", "-c", "ulimit -n"))
	if err != nil {
		t.Fatalf("ulimit: %v", err)
	}

	got, err := strconv.ParseInt(strings.TrimSpace(rr.Stdout.String()), 10, 64)
	if err != nil {
		t.Errorf("ParseInt(%q): %v", rr.Stdout.String(), err)
	}

	// Arbitrary value set by some container runtimes. If higher, apps like MySQL may make bad decisions.
	expected := int64(1048576)
	if got != expected {
		t.Errorf("'ulimit -n' returned %d, expected %d", got, expected)
	}
}

// testPulledImages asserts that this configuration pulls only expected images
func testPulledImages(ctx context.Context, t *testing.T, profile string, version string) {
	t.Helper()
	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), "ssh", "-p", profile, "sudo crictl images -o json"))
	if err != nil {
		t.Errorf("failed tp get images inside minikube. args %q: %v", rr.Command(), err)
	}
	jv := map[string][]struct {
		Tags []string `json:"repoTags"`
	}{}

	stdout := rr.Stdout.String()

	err = json.Unmarshal([]byte(stdout), &jv)
	if err != nil {
		t.Errorf("failed to decode images json %v. output:\n%s", err, stdout)
	}
	found := map[string]bool{}
	for _, img := range jv["images"] {
		for _, i := range img.Tags {
			i = trimImageName(i)
			if defaultImage(i) {
				found[i] = true
			} else {
				t.Logf("Found non-minikube image: %s", i)
			}
		}
	}
	wantRaw, err := images.Kubeadm("", version)
	if err != nil {
		t.Errorf("failed to get kubeadm images for %s : %v", version, err)
	}
	// we need to trim the want raw, because if runtime is docker it will not report the full name with docker.io as prefix
	want := []string{}
	for _, i := range wantRaw {
		want = append(want, trimImageName(i))
	}

	gotImages := []string{}
	for k := range found {
		gotImages = append(gotImages, k)
	}
	sort.Strings(want)
	sort.Strings(gotImages)
	// check if we got all the images we want, ignoring any extraneous ones in cache (eg, may be created by other tests)
	missing := false
	for _, img := range want {
		if sort.SearchStrings(gotImages, img) == len(gotImages) {
			missing = true
			break
		}
	}
	if missing {
		t.Errorf("%s images missing (-want +got):\n%s", version, cmp.Diff(want, gotImages))
	}
}

// testPause asserts that this configuration can be paused and unpaused
func testPause(ctx context.Context, t *testing.T, profile string) {
	t.Helper()
	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), "pause", "-p", profile, "--alsologtostderr", "-v=1"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	got := Status(ctx, t, Target(), profile, "APIServer", profile)
	if got != state.Paused.String() {
		t.Errorf("post-pause apiserver status = %q; want = %q", got, state.Paused)
	}

	got = Status(ctx, t, Target(), profile, "Kubelet", profile)
	if got != state.Stopped.String() {
		t.Errorf("post-pause kubelet status = %q; want = %q", got, state.Stopped)
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "unpause", "-p", profile, "--alsologtostderr", "-v=1"))
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Command(), err)
	}

	got = Status(ctx, t, Target(), profile, "APIServer", profile)
	if got != state.Running.String() {
		t.Errorf("post-unpause apiserver status = %q; want = %q", got, state.Running)
	}

	got = Status(ctx, t, Target(), profile, "Kubelet", profile)
	if got != state.Running.String() {
		t.Errorf("post-unpause kubelet status = %q; want = %q", got, state.Running)
	}

}

// Remove container-specific prefixes for naming consistency
// for example in `docker` runtime we get this:
// 		$ docker@minikube:~$ sudo crictl images -o json | grep dash
// 	         "kubernetesui/dashboard:v2.1.0"
// but for 'containerd' we get full name
// 		$ docker@minikube:~$  sudo crictl images -o json | grep dash
//        	 "docker.io/kubernetesui/dashboard:v2.1.0"
func trimImageName(name string) string {
	name = strings.TrimPrefix(name, "docker.io/")
	name = strings.TrimPrefix(name, "localhost/")
	return name
}

// defaultImage returns true if this image is expected in a default minikube install
func defaultImage(name string) bool {
	if strings.Contains(name, ":latest") {
		return false
	}
	if strings.Contains(name, "k8s.gcr.io") || strings.Contains(name, "kubernetesui") || strings.Contains(name, "storage-provisioner") {
		return true
	}
	return false
}
