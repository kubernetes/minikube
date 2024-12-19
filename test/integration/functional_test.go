//go:build integration

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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/util/retry"

	"github.com/blang/semver/v4"
	"github.com/elazarl/goproxy"
	"github.com/hashicorp/go-retryablehttp"
	cp "github.com/otiai10/copy"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"golang.org/x/build/kubernetes/api"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

const echoServerImg = "kicbase/echo-server"

// validateFunc are for subtests that share a single setup
type validateFunc func(context.Context, *testing.T, string)

// used in validateStartWithProxy and validateSoftStart
var apiPortTest = 8441

// Store the proxy session so we can clean it up at the end
var mitm *StartSession

var runCorpProxy = detect.GithubActionRunner() && runtime.GOOS == "linux" && !arm64Platform()

// TestFunctional are functionality tests which can safely share a profile in parallel
func TestFunctional(t *testing.T) {
	testFunctional(t, "")
}

// TestFunctionalNewestKubernetes are functionality run functional tests using
// NewestKubernetesVersion
func TestFunctionalNewestKubernetes(t *testing.T) {
	if strings.Contains(*startArgs, "--kubernetes-version") || constants.NewestKubernetesVersion == constants.DefaultKubernetesVersion {
		t.Skip()
	}
	k8sVersionString := constants.NewestKubernetesVersion
	t.Run("Version"+k8sVersionString, func(t *testing.T) {
		testFunctional(t, k8sVersionString)
	})

}

func testFunctional(t *testing.T, k8sVersion string) {
	profile := UniqueProfileName("functional")
	ctx := context.WithValue(context.Background(), ContextKey("k8sVersion"), k8sVersion)
	ctx, cancel := context.WithTimeout(ctx, Minutes(40))
	defer func() {
		if !*cleanup {
			return
		}
		p := localSyncTestPath()
		if err := os.Remove(p); err != nil {
			t.Logf("unable to remove %q: %v", p, err)
		}

		Cleanup(t, profile, cancel)
	}()
	// Serial tests
	t.Run("serial", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"CopySyncFile", setupFileSync},                 // Set file for the file sync test case
			{"StartWithProxy", validateStartWithProxy},      // Set everything else up for success
			{"AuditLog", validateAuditAfterStart},           // check audit feature works
			{"SoftStart", validateSoftStart},                // do a soft start. ensure config didn't change.
			{"KubeContext", validateKubeContext},            // Racy: must come immediately after "minikube start"
			{"KubectlGetPods", validateKubectlGetPods},      // Make sure apiserver is up
			{"CacheCmd", validateCacheCmd},                  // Caches images needed for subsequent tests because of proxy
			{"MinikubeKubectlCmd", validateMinikubeKubectl}, // Make sure `minikube kubectl` works
			{"MinikubeKubectlCmdDirectly", validateMinikubeKubectlDirectCall},
			{"ExtraConfig", validateExtraConfig}, // Ensure extra cmdline config change is saved
			{"ComponentHealth", validateComponentHealth},
			{"LogsCmd", validateLogsCmd},
			{"LogsFileCmd", validateLogsFileCmd},
			{"InvalidService", validateInvalidService},
		}
		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			if tc.name == "StartWithProxy" && runCorpProxy {
				tc.name = "StartWithCustomCerts"
				tc.validator = validateStartWithCustomCerts
			}
			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
			})
		}
	})

	defer func() {
		cleanupUnwantedImages(ctx, t, profile)
		if runCorpProxy {
			mitm.Stop(t)
		}
	}()

	// Parallelized tests
	t.Run("parallel", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"ConfigCmd", validateConfigCmd},
			{"DashboardCmd", validateDashboardCmd},
			{"DryRun", validateDryRun},
			{"InternationalLanguage", validateInternationalLanguage},
			{"StatusCmd", validateStatusCmd},
			{"MountCmd", validateMountCmd},
			{"ProfileCmd", validateProfileCmd},
			{"ServiceCmd", validateServiceCmd},
			{"ServiceCmdConnect", validateServiceCmdConnect},
			{"AddonsCmd", validateAddonsCmd},
			{"PersistentVolumeClaim", validatePersistentVolumeClaim},
			{"TunnelCmd", validateTunnelCmd},
			{"SSHCmd", validateSSHCmd},
			{"CpCmd", validateCpCmd},
			{"MySQL", validateMySQL},
			{"FileSync", validateFileSync},
			{"CertSync", validateCertSync},
			{"UpdateContextCmd", validateUpdateContextCmd},
			{"DockerEnv", validateDockerEnv},
			{"PodmanEnv", validatePodmanEnv},
			{"NodeLabels", validateNodeLabels},
			{"ImageCommands", validateImageCommands},
			{"NonActiveRuntimeDisabled", validateNotActiveRuntimeDisabled},
			{"Version", validateVersionCmd},
			{"License", validateLicenseCmd},
		}
		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}

			t.Run(tc.name, func(t *testing.T) {
				MaybeParallel(t)
				tc.validator(ctx, t, profile)
			})
		}
	})

}

func cleanupUnwantedImages(ctx context.Context, t *testing.T, profile string) {
	_, err := exec.LookPath(oci.Docker)
	if err != nil {
		t.Skipf("docker is not installed, cannot delete docker images")
	} else {
		t.Run("delete echo-server images", func(t *testing.T) {
			tags := []string{"1.0", profile}
			for _, tag := range tags {
				image := fmt.Sprintf("%s:%s", echoServerImg, tag)
				rr, err := Run(t, exec.CommandContext(ctx, "docker", "rmi", "-f", image))
				if err != nil {
					t.Logf("failed to remove image %q from docker images. args %q: %v", image, rr.Command(), err)
				}
			}
		})
		t.Run("delete my-image image", func(t *testing.T) {
			newImage := fmt.Sprintf("localhost/my-image:%s", profile)
			rr, err := Run(t, exec.CommandContext(ctx, "docker", "rmi", "-f", newImage))
			if err != nil {
				t.Logf("failed to remove image my-image from docker images. args %q: %v", rr.Command(), err)
			}
		})

		t.Run("delete minikube cached images", func(t *testing.T) {
			img := "minikube-local-cache-test:" + profile
			rr, err := Run(t, exec.CommandContext(ctx, "docker", "rmi", "-f", img))
			if err != nil {
				t.Logf("failed to remove image minikube local cache test images from docker. args %q: %v", rr.Command(), err)
			}
		})
	}
}

// validateNodeLabels checks if minikube cluster is created with correct kubernetes's node label
func validateNodeLabels(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: Get the node labels from the cluster with `kubectl get nodes`
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "nodes", "--output=go-template", "--template='{{range $k, $v := (index .items 0).metadata.labels}}{{$k}} {{end}}'"))
	if err != nil {
		t.Errorf("failed to 'kubectl get nodes' with args %q: %v", rr.Command(), err)
	}
	// docs: check if the node labels matches with the expected Minikube labels: `minikube.k8s.io/*`
	expectedLabels := []string{"minikube.k8s.io/commit", "minikube.k8s.io/version", "minikube.k8s.io/updated_at", "minikube.k8s.io/name", "minikube.k8s.io/primary"}
	for _, el := range expectedLabels {
		if !strings.Contains(rr.Output(), el) {
			t.Errorf("expected to have label %q in node labels but got : %s", el, rr.Output())
		}
	}
}

// tagAndLoadImage is a helper function to pull, tag, load image (decreases cyclomatic complexity for linter).
func tagAndLoadImage(ctx context.Context, t *testing.T, profile, taggedImage string) {
	newPulledImage := fmt.Sprintf("%s:%s", echoServerImg, "latest")
	rr, err := Run(t, exec.CommandContext(ctx, "docker", "pull", newPulledImage))
	if err != nil {
		t.Fatalf("failed to setup test (pull image): %v\n%s", err, rr.Output())
	}

	rr, err = Run(t, exec.CommandContext(ctx, "docker", "tag", newPulledImage, taggedImage))
	if err != nil {
		t.Fatalf("failed to setup test (tag image) : %v\n%s", err, rr.Output())
	}

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "load", "--daemon", taggedImage, "--alsologtostderr"))
	if err != nil {
		t.Fatalf("loading image into minikube from daemon: %v\n%s", err, rr.Output())
	}

	checkImageExists(ctx, t, profile, taggedImage)
}

// runImageList is a helper function to run 'image ls' command test.
func runImageList(ctx context.Context, t *testing.T, profile, testName, format, expectedFormat string) {
	expectedResult := expectedImageFormat(expectedFormat)

	// docs: Make sure image listing works by `minikube image ls`
	t.Run(testName, func(t *testing.T) {
		MaybeParallel(t)

		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "ls", "--format", format, "--alsologtostderr"))
		if err != nil {
			t.Fatalf("listing image with minikube: %v\n%s", err, rr.Output())
		}
		if rr.Stdout.Len() > 0 {
			t.Logf("(dbg) Stdout: %s:\n%s", rr.Command(), rr.Stdout)
		}
		if rr.Stderr.Len() > 0 {
			t.Logf("(dbg) Stderr: %s:\n%s", rr.Command(), rr.Stderr)
		}

		list := rr.Output()
		for _, theImage := range expectedResult {
			if !strings.Contains(list, theImage) {
				t.Fatalf("expected %s to be listed with minikube but the image is not there", theImage)
			}
		}
	})
}

func expectedImageFormat(format string) []string {
	return []string{
		fmt.Sprintf(format, "registry.k8s.io/pause"),
		fmt.Sprintf(format, "registry.k8s.io/kube-apiserver"),
	}
}

// validateImageCommands runs tests on all the `minikube image` commands, ex. `minikube image load`, `minikube image list`, etc.
func validateImageCommands(ctx context.Context, t *testing.T, profile string) {
	// docs(skip): Skips on `none` driver as image loading is not supported
	if NoneDriver() {
		t.Skip("image commands are not available on the none driver")
	}
	// docs(skip): Skips on GitHub Actions and macOS as this test case requires a running docker daemon
	if detect.GithubActionRunner() && runtime.GOOS == "darwin" {
		t.Skip("skipping on darwin github action runners, as this test requires a running docker daemon")
	}

	runImageList(ctx, t, profile, "ImageListShort", "short", "%s")
	runImageList(ctx, t, profile, "ImageListTable", "table", "| %s")
	runImageList(ctx, t, profile, "ImageListJson", "json", "[\"%s")
	runImageList(ctx, t, profile, "ImageListYaml", "yaml", "- %s")

	// docs: Make sure image building works by `minikube image build`
	t.Run("ImageBuild", func(t *testing.T) {
		MaybeParallel(t)

		if _, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "pgrep", "buildkitd")); err == nil {
			t.Errorf("buildkitd process is running, should not be running until `minikube image build` is ran")
		}

		newImage := fmt.Sprintf("localhost/my-image:%s", profile)

		// try to build the new image with minikube
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "build", "-t", newImage, filepath.Join(*testdataDir, "build"), "--alsologtostderr"))
		if err != nil {
			t.Fatalf("building image with minikube: %v\n%s", err, rr.Output())
		}
		if rr.Stdout.Len() > 0 {
			t.Logf("(dbg) Stdout: %s:\n%s", rr.Command(), rr.Stdout)
		}
		if rr.Stderr.Len() > 0 {
			t.Logf("(dbg) Stderr: %s:\n%s", rr.Command(), rr.Stderr)
		}

		checkImageExists(ctx, t, profile, newImage)
	})

	taggedImage := fmt.Sprintf("%s:%s", echoServerImg, profile)
	imageFile := "echo-server-save.tar"
	var imagePath string
	defer os.Remove(imageFile)

	t.Run("Setup", func(t *testing.T) {
		var err error
		imagePath, err = filepath.Abs(imageFile)
		if err != nil {
			t.Fatalf("failed to get absolute path of file %q: %v", imageFile, err)
		}

		pulledImage := fmt.Sprintf("%s:%s", echoServerImg, "1.0")
		rr, err := Run(t, exec.CommandContext(ctx, "docker", "pull", pulledImage))
		if err != nil {
			t.Fatalf("failed to setup test (pull image): %v\n%s", err, rr.Output())
		}

		rr, err = Run(t, exec.CommandContext(ctx, "docker", "tag", pulledImage, taggedImage))
		if err != nil {
			t.Fatalf("failed to setup test (tag image) : %v\n%s", err, rr.Output())
		}
	})

	// docs: Make sure image loading from Docker daemon works by `minikube image load --daemon`
	t.Run("ImageLoadDaemon", func(t *testing.T) {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "load", "--daemon", taggedImage, "--alsologtostderr"))
		if err != nil {
			t.Fatalf("loading image into minikube from daemon: %v\n%s", err, rr.Output())
		}

		checkImageExists(ctx, t, profile, taggedImage)
	})

	// docs: Try to load image already loaded and make sure `minikube image load --daemon` works
	t.Run("ImageReloadDaemon", func(t *testing.T) {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "load", "--daemon", taggedImage, "--alsologtostderr"))
		if err != nil {
			t.Fatalf("loading image into minikube from daemon: %v\n%s", err, rr.Output())
		}

		checkImageExists(ctx, t, profile, taggedImage)
	})

	// docs: Make sure a new updated tag works by `minikube image load --daemon`
	t.Run("ImageTagAndLoadDaemon", func(t *testing.T) {
		tagAndLoadImage(ctx, t, profile, taggedImage)
	})

	// docs: Make sure image saving works by `minikube image load --daemon`
	t.Run("ImageSaveToFile", func(t *testing.T) {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "save", taggedImage, imagePath, "--alsologtostderr"))
		if err != nil {
			t.Fatalf("saving image from minikube to file: %v\n%s", err, rr.Output())
		}

		if _, err := os.Stat(imagePath); err != nil {
			t.Fatalf("expected %q to exist after `image save`, but doesn't exist", imagePath)
		}
	})

	// docs: Make sure image removal works by `minikube image rm`
	t.Run("ImageRemove", func(t *testing.T) {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "rm", taggedImage, "--alsologtostderr"))
		if err != nil {
			t.Fatalf("removing image from minikube: %v\n%s", err, rr.Output())
		}

		// make sure the image was removed
		rr, err = listImages(ctx, t, profile)
		if err != nil {
			t.Fatalf("listing images: %v\n%s", err, rr.Output())
		}
		if strings.Contains(rr.Output(), taggedImage) {
			t.Fatalf("expected %q to be removed from minikube but still exists", taggedImage)
		}
	})

	// docs: Make sure image loading from file works by `minikube image load`
	t.Run("ImageLoadFromFile", func(t *testing.T) {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "load", imagePath, "--alsologtostderr"))
		if err != nil || strings.Contains(rr.Output(), "failed pushing to: functional") {
			t.Fatalf("loading image into minikube from file: %v\n%s", err, rr.Output())
		}

		checkImageExists(ctx, t, profile, taggedImage)
	})

	// docs: Make sure image saving to Docker daemon works by `minikube image load`
	t.Run("ImageSaveDaemon", func(t *testing.T) {
		rr, err := Run(t, exec.CommandContext(ctx, "docker", "rmi", taggedImage))
		if err != nil {
			t.Fatalf("failed to remove image from docker: %v\n%s", err, rr.Output())
		}

		rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "save", "--daemon", taggedImage, "--alsologtostderr"))
		if err != nil {
			t.Fatalf("saving image from minikube to daemon: %v\n%s", err, rr.Output())
		}
		imageToDelete := taggedImage
		if ContainerRuntime() == "crio" {
			imageToDelete = cruntime.AddLocalhostPrefix(imageToDelete)
		}
		rr, err = Run(t, exec.CommandContext(ctx, "docker", "image", "inspect", imageToDelete))
		if err != nil {
			t.Fatalf("expected image to be loaded into Docker, but image was not found: %v\n%s", err, rr.Output())
		}
	})
}

func checkImageExists(ctx context.Context, t *testing.T, profile string, image string) {
	// make sure the image was correctly loaded
	rr, err := listImages(ctx, t, profile)
	if err != nil {
		t.Fatalf("listing images: %v\n%s", err, rr.Output())
	}
	if !strings.Contains(rr.Output(), image) {
		t.Fatalf("expected %q to be loaded into minikube but the image is not there", image)
	}
}

func listImages(ctx context.Context, t *testing.T, profile string) (*RunResult, error) {
	return Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "image", "ls"))
}

// check functionality of minikube after evaluating docker-env
func validateDockerEnv(ctx context.Context, t *testing.T, profile string) {
	// docs(skip): Skips on `none` drive since `docker-env` is not supported
	if NoneDriver() {
		t.Skipf("none driver does not support docker-env")
	}

	// docs(skip): Skips on non-docker container runtime
	if cr := ContainerRuntime(); cr != "docker" {
		t.Skipf("only validate docker env with docker container runtime, currently testing %s", cr)
	}
	defer PostMortemLogs(t, profile)

	type ShellTest struct {
		name          string
		commandPrefix []string
		formatArg     string
	}

	// docs: Run `eval $(minikube docker-env)` to configure current environment to use minikube's Docker daemon
	windowsTests := []ShellTest{
		{"powershell", []string{"powershell.exe", "-NoProfile", "-NonInteractive"}, "%[1]s -p %[2]s docker-env | Invoke-Expression ; "},
	}
	posixTests := []ShellTest{
		{"bash", []string{"/bin/bash", "-c"}, "eval $(%[1]s -p %[2]s docker-env) && "},
	}

	tests := posixTests
	if runtime.GOOS == "windows" {
		tests = windowsTests
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mctx, cancel := context.WithTimeout(ctx, Seconds(120))
			defer cancel()

			command := make([]string, len(tc.commandPrefix)+1)
			copy(command, tc.commandPrefix)

			formattedArg := fmt.Sprintf(tc.formatArg, Target(), profile)

			// docs: Run `minikube status` to get the minikube status
			// we should be able to get minikube status with a shell which evaled docker-env
			command[len(command)-1] = formattedArg + Target() + " status -p " + profile
			c := exec.CommandContext(mctx, command[0], command[1:]...)
			rr, err := Run(t, c)

			if mctx.Err() == context.DeadlineExceeded {
				t.Errorf("failed to run the command by deadline. exceeded timeout. %s", rr.Command())
			}
			if err != nil {
				t.Fatalf("failed to do status after eval-ing docker-env. error: %v", err)
			}
			// docs: Make sure minikube components have status `Running`
			if !strings.Contains(rr.Output(), "Running") {
				t.Fatalf("expected status output to include 'Running' after eval docker-env but got: *%s*", rr.Output())
			}
			// docs: Make sure `docker-env` has status `in-use`
			if !strings.Contains(rr.Output(), "in-use") {
				t.Fatalf("expected status output to include `in-use` after eval docker-env but got *%s*", rr.Output())
			}

			mctx, cancel = context.WithTimeout(ctx, Seconds(60))
			defer cancel()

			// docs: Run eval `$(minikube -p profile docker-env)` and check if we are point to docker inside minikube
			command[len(command)-1] = formattedArg + "docker images"
			c = exec.CommandContext(mctx, command[0], command[1:]...)
			rr, err = Run(t, c)

			if mctx.Err() == context.DeadlineExceeded {
				t.Errorf("failed to run the command in 30 seconds. exceeded 30s timeout. %s", rr.Command())
			}

			if err != nil {
				t.Fatalf("failed to run minikube docker-env. args %q : %v ", rr.Command(), err)
			}

			// docs: Make sure `docker images` hits the minikube's Docker daemon by check if `gcr.io/k8s-minikube/storage-provisioner` is in the output of `docker images`
			expectedImgInside := "gcr.io/k8s-minikube/storage-provisioner"
			if !strings.Contains(rr.Output(), expectedImgInside) {
				t.Fatalf("expected 'docker images' to have %q inside minikube. but the output is: *%s*", expectedImgInside, rr.Output())
			}
		})
	}
}

// check functionality of minikube after evaluating podman-env
func validatePodmanEnv(ctx context.Context, t *testing.T, profile string) {
	// docs(skip): Skips on `none` drive since `podman-env` is not supported
	if NoneDriver() {
		t.Skipf("none driver does not support podman-env")
	}

	// docs(skip): Skips on non-docker container runtime
	if cr := ContainerRuntime(); cr != "podman" {
		t.Skipf("only validate podman env with docker container runtime, currently testing %s", cr)
	}

	// docs(skip): Skips on non-Linux platforms
	if runtime.GOOS != "linux" {
		t.Skipf("only validate podman env on linux, currently testing %s", runtime.GOOS)
	}

	defer PostMortemLogs(t, profile)

	mctx, cancel := context.WithTimeout(ctx, Seconds(120))
	defer cancel()

	// docs: Run `eval $(minikube podman-env)` to configure current environment to use minikube's Podman daemon, and `minikube status` to get the minikube status
	c := exec.CommandContext(mctx, "/bin/bash", "-c", "eval $("+Target()+" -p "+profile+" podman-env) && "+Target()+" status -p "+profile)
	// we should be able to get minikube status with a bash which evaluated podman-env
	rr, err := Run(t, c)

	if mctx.Err() == context.DeadlineExceeded {
		t.Errorf("failed to run the command by deadline. exceeded timeout. %s", rr.Command())
	}
	if err != nil {
		t.Fatalf("failed to do status after eval-ing podman-env. error: %v", err)
	}
	// docs: Make sure minikube components have status `Running`
	if !strings.Contains(rr.Output(), "Running") {
		t.Fatalf("expected status output to include 'Running' after eval podman-env but got: *%s*", rr.Output())
	}
	// docs: Make sure `podman-env` has status `in-use`
	if !strings.Contains(rr.Output(), "in-use") {
		t.Fatalf("expected status output to include `in-use` after eval podman-env but got *%s*", rr.Output())
	}

	mctx, cancel = context.WithTimeout(ctx, Seconds(60))
	defer cancel()
	// docs: Run `eval $(minikube docker-env)` again and `docker images` to list the docker images using the minikube's Docker daemon
	c = exec.CommandContext(mctx, "/bin/bash", "-c", "eval $("+Target()+" -p "+profile+" podman-env) && docker images")
	rr, err = Run(t, c)

	if mctx.Err() == context.DeadlineExceeded {
		t.Errorf("failed to run the command in 30 seconds. exceeded 30s timeout. %s", rr.Command())
	}
	if err != nil {
		t.Fatalf("failed to run minikube podman-env. args %q : %v ", rr.Command(), err)
	}

	// docs: Make sure `docker images` hits the minikube's Podman daemon by check if `gcr.io/k8s-minikube/storage-provisioner` is in the output of `docker images`
	expectedImgInside := "gcr.io/k8s-minikube/storage-provisioner"
	if !strings.Contains(rr.Output(), expectedImgInside) {
		t.Fatalf("expected 'docker images' to have %q inside minikube. but the output is: *%s*", expectedImgInside, rr.Output())
	}
}

// validateStartWithProxy makes sure minikube start respects the HTTP_PROXY environment variable
func validateStartWithProxy(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	// docs: Start a local HTTP proxy
	srv, err := startHTTPProxy(t)
	if err != nil {
		t.Fatalf("failed to set up the test proxy: %s", err)
	}

	// docs: Start minikube with the environment variable `HTTP_PROXY` set to the local HTTP proxy
	startMinikubeWithProxy(ctx, t, profile, "HTTP_PROXY", srv.Addr)
}

// validateStartWithCustomCerts makes sure minikube start respects the HTTPS_PROXY environment variable and works with custom certs
// a proxy is started by calling the mitmdump binary in the background, then installing the certs generated by the binary
// mitmproxy/dump creates the proxy at localhost at port 8080
// only runs on GitHub Actions for amd64 linux, otherwise validateStartWithProxy runs instead
func validateStartWithCustomCerts(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	err := startProxyWithCustomCerts(ctx, t)
	if err != nil {
		t.Fatalf("failed to set up the test proxy: %s", err)
	}

	startMinikubeWithProxy(ctx, t, profile, "HTTPS_PROXY", "127.0.0.1:8080")
}

// validateAuditAfterStart makes sure the audit log contains the correct logging after minikube start
func validateAuditAfterStart(_ context.Context, t *testing.T, profile string) {
	// docs: Read the audit log file and make sure it contains the current minikube profile name
	got, err := auditContains(profile)
	if err != nil {
		t.Fatalf("failed to check audit log: %v", err)
	}
	if !got {
		t.Errorf("audit.json does not contain the profile %q", profile)
	}
}

// validateSoftStart validates that after minikube already started, a `minikube start` should not change the configs.
func validateSoftStart(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	start := time.Now()
	// docs: The test `validateStartWithProxy` should have start minikube, make sure the configured node port is `8441`
	beforeCfg, err := config.LoadProfile(profile)
	if err != nil {
		t.Fatalf("error reading cluster config before soft start: %v", err)
	}
	if beforeCfg.Config.APIServerPort != apiPortTest {
		t.Errorf("expected cluster config node port before soft start to be %d but got %d", apiPortTest, beforeCfg.Config.APIServerPort)
	}

	// docs: Run `minikube start` again as a soft start
	softStartArgs := []string{"start", "-p", profile, "--alsologtostderr", "-v=8"}
	c := exec.CommandContext(ctx, Target(), softStartArgs...)
	rr, err := Run(t, c)
	if err != nil {
		t.Errorf("failed to soft start minikube. args %q: %v", rr.Command(), err)
	}
	t.Logf("soft start took %s for %q cluster.", time.Since(start), profile)

	// docs: Make sure the configured node port is not changed
	afterCfg, err := config.LoadProfile(profile)
	if err != nil {
		t.Errorf("error reading cluster config after soft start: %v", err)
	}

	if afterCfg.Config.APIServerPort != apiPortTest {
		t.Errorf("expected node port in the config not to change after soft start. expected node port to be %d but got %d.", apiPortTest, afterCfg.Config.APIServerPort)
	}
}

// validateKubeContext asserts that kubectl is properly configured (race-condition prone!)
func validateKubeContext(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: Run `kubectl config current-context`
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "config", "current-context"))
	if err != nil {
		t.Errorf("failed to get current-context. args %q : %v", rr.Command(), err)
	}
	// docs: Make sure the current minikube profile name is in the output of the command
	if !strings.Contains(rr.Stdout.String(), profile) {
		t.Errorf("expected current-context = %q, but got *%q*", profile, rr.Stdout.String())
	}
}

// validateKubectlGetPods asserts that `kubectl get pod -A` returns non-zero content
func validateKubectlGetPods(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: Run `kubectl get po -A` to get all pods in the current minikube profile
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "po", "-A"))
	if err != nil {
		t.Errorf("failed to get kubectl pods: args %q : %v", rr.Command(), err)
	}
	// docs: Make sure the output is not empty and contains `kube-system` components
	if rr.Stderr.String() != "" {
		t.Errorf("expected stderr to be empty but got *%q*: args %q", rr.Stderr, rr.Command())
	}
	if !strings.Contains(rr.Stdout.String(), "kube-system") {
		t.Errorf("expected stdout to include *kube-system* but got *%q*. args: %q", rr.Stdout, rr.Command())
	}
}

// validateMinikubeKubectl validates that the `minikube kubectl` command returns content
func validateMinikubeKubectl(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: Run `minikube kubectl -- get pods` to get the pods in the current minikube profile
	// Must set the profile so that it knows what version of Kubernetes to use
	kubectlArgs := []string{"-p", profile, "kubectl", "--", "--context", profile, "get", "pods"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), kubectlArgs...))
	// docs: Make sure the command doesn't raise any error
	if err != nil {
		t.Fatalf("failed to get pods. args %q: %v", rr.Command(), err)
	}
}

// validateMinikubeKubectlDirectCall validates that calling minikube's kubectl
func validateMinikubeKubectlDirectCall(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	dir := filepath.Dir(Target())
	newName := "kubectl"
	if runtime.GOOS == "windows" {
		newName += ".exe"
	}
	dstfn := filepath.Join(dir, newName)
	err := os.Link(Target(), dstfn)

	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dstfn) // clean up

	// docs: Run `kubectl get pods` by calling the minikube's `kubectl` binary file directly
	kubectlArgs := []string{"--context", profile, "get", "pods"}
	rr, err := Run(t, exec.CommandContext(ctx, dstfn, kubectlArgs...))
	// docs: Make sure the command doesn't raise any error
	if err != nil {
		t.Fatalf("failed to run kubectl directly. args %q: %v", rr.Command(), err)
	}
}

// validateExtraConfig verifies minikube with --extra-config works as expected
func validateExtraConfig(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	start := time.Now()
	// docs: The tests before this already created a profile
	// docs: Soft-start minikube with different `--extra-config` command line option
	startArgs := []string{"start", "-p", profile, "--extra-config=apiserver.enable-admission-plugins=NamespaceAutoProvision", "--wait=all"}
	c := exec.CommandContext(ctx, Target(), startArgs...)
	rr, err := Run(t, c)
	if err != nil {
		t.Errorf("failed to restart minikube. args %q: %v", rr.Command(), err)
	}
	t.Logf("restart took %s for %q cluster.", time.Since(start), profile)

	// docs: Load the profile's config
	afterCfg, err := config.LoadProfile(profile)
	if err != nil {
		t.Errorf("error reading cluster config after soft start: %v", err)
	}

	// docs: Make sure the specified `--extra-config` is correctly returned
	expectedExtraOptions := "apiserver.enable-admission-plugins=NamespaceAutoProvision"
	if !strings.Contains(afterCfg.Config.KubernetesConfig.ExtraOptions.String(), expectedExtraOptions) {
		t.Errorf("expected ExtraOptions to contain %s but got %s", expectedExtraOptions, afterCfg.Config.KubernetesConfig.ExtraOptions.String())
	}
}

// imageID returns a docker image id for image `image` and current architecture
// 'image' is supposed to be one commonly used in minikube integration tests,
// like k8s 'pause'
func imageID(image string) string {
	ids := map[string]map[string]string{
		"pause": {
			"amd64": "0184c1613d929",
			"arm64": "3d18732f8686c",
		},
	}

	if imgIDs, ok := ids[image]; ok {
		if id, ok := imgIDs[runtime.GOARCH]; ok {
			return id
		}
		panic(fmt.Sprintf("unexpected architecture for image %q: %v", image, runtime.GOARCH))
	}
	panic("unexpected image name: " + image)
}

// validateComponentHealth asserts that all Kubernetes components are healthy
// NOTE: It expects all components to be Ready, so it makes sense to run it close after only those tests that include '--wait=all' start flag
func validateComponentHealth(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// The ComponentStatus API is deprecated in v1.19, so do the next closest thing.
	found := map[string]bool{
		"etcd":                    false,
		"kube-apiserver":          false,
		"kube-controller-manager": false,
		"kube-scheduler":          false,
	}

	// docs: Run `kubectl get po po -l tier=control-plane -n kube-system -o=json` to get all the Kubernetes conponents
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "po", "-l", "tier=control-plane", "-n", "kube-system", "-o=json"))
	if err != nil {
		t.Fatalf("failed to get components. args %q: %v", rr.Command(), err)
	}
	cs := api.PodList{}
	d := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes()))
	if err := d.Decode(&cs); err != nil {
		t.Fatalf("failed to decode kubectl json output: args %q : %v", rr.Command(), err)
	}

	// docs: For each component, make sure the pod status is `Running`
	for _, i := range cs.Items {
		for _, l := range i.Labels {
			if _, ok := found[l]; ok { // skip irrelevant (eg, repeating/redundant '"tier": "control-plane"') labels
				found[l] = true
				t.Logf("%s phase: %s", l, i.Status.Phase)
				if i.Status.Phase != api.PodRunning {
					t.Errorf("%s is not Running: %+v", l, i.Status)
					continue
				}
				for _, c := range i.Status.Conditions {
					if c.Type == api.PodReady {
						if c.Status != api.ConditionTrue {
							t.Errorf("%s is not Ready: %+v", l, i.Status)
						} else {
							t.Logf("%s status: %s", l, c.Type)
						}
						break
					}
				}
			}
		}
	}

	for k, v := range found {
		if !v {
			t.Errorf("expected component %q was not found", k)
		}
	}
}

// validateStatusCmd makes sure `minikube status` outputs correctly
func validateStatusCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status"))
	if err != nil {
		t.Errorf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	// docs: Run `minikube status` with custom format `host:{{.Host}},kublet:{{.Kubelet}},apiserver:{{.APIServer}},kubeconfig:{{.Kubeconfig}}`
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-f", "host:{{.Host}},kublet:{{.Kubelet}},apiserver:{{.APIServer}},kubeconfig:{{.Kubeconfig}}"))
	if err != nil {
		t.Errorf("failed to run minikube status with custom format: args %q: %v", rr.Command(), err)
	}
	// docs: Make sure `host`, `kublete`, `apiserver` and `kubeconfig` statuses are shown in the output
	re := `host:([A-z]+),kublet:([A-z]+),apiserver:([A-z]+),kubeconfig:([A-z]+)`
	match, _ := regexp.MatchString(re, rr.Stdout.String())
	if !match {
		t.Errorf("failed to match regex %q for minikube status with custom format. args %q. output: %s", re, rr.Command(), rr.Output())
	}

	// docs: Run `minikube status` again as JSON output
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-o", "json"))
	if err != nil {
		t.Errorf("failed to run minikube status with json output. args %q : %v", rr.Command(), err)
	}
	// docs: Make sure `host`, `kublete`, `apiserver` and `kubeconfig` statuses are set in the JSON output
	var jsonObject map[string]interface{}
	err = json.Unmarshal(rr.Stdout.Bytes(), &jsonObject)
	if err != nil {
		t.Errorf("failed to decode json from minikube status. args %q. %v", rr.Command(), err)
	}
	if _, ok := jsonObject["Host"]; !ok {
		t.Errorf("%q failed: %v. Missing key %s in json object", rr.Command(), err, "Host")
	}
	if _, ok := jsonObject["Kubelet"]; !ok {
		t.Errorf("%q failed: %v. Missing key %s in json object", rr.Command(), err, "Kubelet")
	}
	if _, ok := jsonObject["APIServer"]; !ok {
		t.Errorf("%q failed: %v. Missing key %s in json object", rr.Command(), err, "APIServer")
	}
	if _, ok := jsonObject["Kubeconfig"]; !ok {
		t.Errorf("%q failed: %v. Missing key %s in json object", rr.Command(), err, "Kubeconfig")
	}
}

// validateDashboardCmd asserts that the dashboard command works
func validateDashboardCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	mctx, cancel := context.WithTimeout(ctx, Seconds(300))
	defer cancel()

	// docs: Run `minikube dashboard --url` to start minikube dashboard and return the URL of it
	args := []string{"dashboard", "--url", "--port", "36195", "-p", profile, "--alsologtostderr", "-v=1"}
	ss, err := Start(t, exec.CommandContext(mctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to run minikube dashboard. args %q : %v", args, err)
	}
	defer func() {
		ss.Stop(t)
	}()

	s, err := dashboardURL(ss.Stdout)
	if err != nil {
		if runtime.GOOS == "windows" {
			t.Skip(err)
		}
		t.Fatal(err)
	}

	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		t.Fatalf("failed to parse %q: %v", s, err)
	}

	// docs: Send a GET request to the dashboard URL
	resp, err := retryablehttp.Get(u.String())
	if err != nil {
		t.Fatalf("failed to http get %q: %v\nresponse: %+v", u.String(), err, resp)
	}

	// docs: Make sure HTTP status OK is returned
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("failed to read http response body from dashboard %q: %v", u.String(), err)
		}
		t.Errorf("%s returned status code %d, expected %d.\nbody:\n%s", u, resp.StatusCode, http.StatusOK, body)
	}
}

// dashboardURL gets the dashboard URL from the command stdout.
func dashboardURL(b *bufio.Reader) (string, error) {
	// match http://127.0.0.1:XXXXX/api/v1/namespaces/kubernetes-dashboard/services/http:kubernetes-dashboard:/proxy/
	dashURLRegexp := regexp.MustCompile(`^http:\/\/127\.0\.0\.1:[0-9]{5}\/api\/v1\/namespaces\/kubernetes-dashboard\/services\/http:kubernetes-dashboard:\/proxy\/$`)

	s := bufio.NewScanner(b)
	for s.Scan() {
		t := s.Text()
		if dashURLRegexp.MatchString(t) {
			return t, nil
		}
	}
	if err := s.Err(); err != nil {
		return "", fmt.Errorf("failed reading input: %v", err)
	}
	return "", fmt.Errorf("output didn't produce a URL")
}

// validateDryRun asserts that the dry-run mode quickly exits with the right code
func validateDryRun(ctx context.Context, t *testing.T, profile string) {
	// dry-run mode should always be able to finish quickly (<5s) expect Docker Windows
	timeout := Seconds(5)
	if runtime.GOOS == "windows" && DockerDriver() {
		timeout = Seconds(10)
	}
	mctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// docs: Run `minikube start --dry-run --memory 250MB`
	// Too little memory!
	startArgs := append([]string{"start", "-p", profile, "--dry-run", "--memory", "250MB", "--alsologtostderr"}, StartArgsWithContext(ctx)...)
	c := exec.CommandContext(mctx, Target(), startArgs...)
	rr, err := Run(t, c)

	// docs: Since the 250MB memory is less than the required 2GB, minikube should exit with an exit code `ExInsufficientMemory`
	wantCode := reason.ExInsufficientMemory
	if rr.ExitCode != wantCode {
		if HyperVDriver() {
			t.Skip("skipping this error on HyperV till this issue is solved https://github.com/kubernetes/minikube/issues/9785")
		} else {
			t.Errorf("dry-run(250MB) exit code = %d, wanted = %d: %v", rr.ExitCode, wantCode, err)
		}
	}

	dctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// docs: Run `minikube start --dry-run`
	startArgs = append([]string{"start", "-p", profile, "--dry-run", "--alsologtostderr", "-v=1"}, StartArgsWithContext(ctx)...)
	c = exec.CommandContext(dctx, Target(), startArgs...)
	rr, err = Run(t, c)
	// docs: Make sure the command doesn't raise any error
	if rr.ExitCode != 0 || err != nil {
		if HyperVDriver() {
			t.Skip("skipping this error on HyperV till this issue is solved https://github.com/kubernetes/minikube/issues/9785")
		} else {
			t.Errorf("dry-run exit code = %d, wanted = %d: %v", rr.ExitCode, 0, err)
		}

	}
}

// validateInternationalLanguage asserts that the language used can be changed with environment variables
func validateInternationalLanguage(ctx context.Context, t *testing.T, profile string) {
	// dry-run mode should always be able to finish quickly (<5s) except Docker Windows
	timeout := Seconds(5)
	if runtime.GOOS == "windows" && DockerDriver() {
		timeout = Seconds(10)
	}
	mctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Too little memory!
	startArgs := append([]string{"start", "-p", profile, "--dry-run", "--memory", "250MB", "--alsologtostderr"}, StartArgsWithContext(ctx)...)
	c := exec.CommandContext(mctx, Target(), startArgs...)
	// docs: Set environment variable `LC_ALL=fr` to enable minikube translation to French
	c.Env = append(os.Environ(), "LC_ALL=fr")

	// docs: Start minikube with memory of 250MB which is too little: `minikube start --dry-run --memory 250MB`
	rr, err := Run(t, c)

	wantCode := reason.ExInsufficientMemory
	if rr.ExitCode != wantCode {
		if HyperVDriver() {
			t.Skip("skipping this error on HyperV till this issue is solved https://github.com/kubernetes/minikube/issues/9785")
		} else {
			t.Errorf("dry-run(250MB) exit code = %d, wanted = %d: %v", rr.ExitCode, wantCode, err)
		}
	}
	// docs: Make sure the dry-run output message is in French
	if !strings.Contains(rr.Stdout.String(), "Utilisation du pilote") {
		t.Errorf("dry-run output was expected to be in French. Expected \"Utilisation du pilote\", but not present in output:\n%s", rr.Stdout.String())
	}
}

// validateCacheCmd tests functionality of cache command (cache add, delete, list)
func validateCacheCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	if NoneDriver() {
		t.Skipf("skipping: cache unsupported by none")
	}

	t.Run("cache", func(t *testing.T) {

		// docs: Run `minikube cache add` and make sure we can add a remote image to the cache
		t.Run("add_remote", func(t *testing.T) {
			for _, img := range []string{"registry.k8s.io/pause:3.1", "registry.k8s.io/pause:3.3", "registry.k8s.io/pause:latest"} {
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "cache", "add", img))
				if err != nil {
					t.Errorf("failed to 'cache add' remote image %q. args %q err %v", img, rr.Command(), err)
				}
			}
		})

		// docs: Run `minikube cache add` and make sure we can build and add a local image to the cache
		t.Run("add_local", func(t *testing.T) {
			if detect.GithubActionRunner() && runtime.GOOS == "darwin" {
				t.Skipf("skipping this test because Docker can not run in macos on github action free version. https://github.community/t/is-it-possible-to-install-and-configure-docker-on-macos-runner/16981")
			}

			_, err := exec.LookPath(oci.Docker)
			if err != nil {
				t.Skipf("docker is not installed, skipping local image test")
			}

			dname := t.TempDir()

			message := []byte("FROM scratch\nADD Dockerfile /x")
			err = os.WriteFile(filepath.Join(dname, "Dockerfile"), message, 0644)
			if err != nil {
				t.Fatalf("unable to write Dockerfile: %v", err)
			}

			img := "minikube-local-cache-test:" + profile

			_, err = Run(t, exec.CommandContext(ctx, "docker", "build", "-t", img, dname))
			if err != nil {
				t.Skipf("failed to build docker image, skipping local test: %v", err)
			}

			defer func() {
				_, err := Run(t, exec.CommandContext(ctx, "docker", "rmi", img))
				if err != nil {
					t.Errorf("failed to delete local image %q, err %v", img, err)
				}
			}()

			rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "cache", "add", img))
			if err != nil {
				t.Errorf("failed to 'cache add' local image %q. args %q err %v", img, rr.Command(), err)
			}

			rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "cache", "delete", img))
			if err != nil {
				t.Errorf("failed to 'cache delete' local image %q. args %q err %v", img, rr.Command(), err)
			}
		})

		// docs: Run `minikube cache delete` and make sure we can delete an image from the cache
		t.Run("CacheDelete", func(t *testing.T) {
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "cache", "delete", "registry.k8s.io/pause:3.3"))
			if err != nil {
				t.Errorf("failed to delete image registry.k8s.io/pause:3.3 from cache. args %q: %v", rr.Command(), err)
			}
		})

		// docs: Run `minikube cache list` and make sure we can list the images in the cache
		t.Run("list", func(t *testing.T) {
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "cache", "list"))
			if err != nil {
				t.Errorf("failed to do cache list. args %q: %v", rr.Command(), err)
			}
			if !strings.Contains(rr.Output(), "registry.k8s.io/pause") {
				t.Errorf("expected 'cache list' output to include 'registry.k8s.io/pause' but got: ***%s***", rr.Output())
			}
			if strings.Contains(rr.Output(), "registry.k8s.io/pause:3.3") {
				t.Errorf("expected 'cache list' output not to include registry.k8s.io/pause:3.3 but got: ***%s***", rr.Output())
			}
		})

		// docs: Run `minikube ssh sudo crictl images` and make sure we can list the images in the cache with `crictl`
		t.Run("verify_cache_inside_node", func(t *testing.T) {
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "sudo", "crictl", "images"))
			if err != nil {
				t.Errorf("failed to get images by %q ssh %v", rr.Command(), err)
			}
			pauseID := imageID("pause")
			if !strings.Contains(rr.Output(), pauseID) {
				t.Errorf("expected sha for pause:3.3 %q to be in the output but got *%s*", pauseID, rr.Output())
			}
		})

		// docs: Delete an image from minikube node and run `minikube cache reload` to make sure the image is brought back correctly
		t.Run("cache_reload", func(t *testing.T) { // deleting image inside minikube node manually and expecting reload to bring it back
			img := "registry.k8s.io/pause:latest"
			// deleting image inside minikube node manually

			var binary string
			switch ContainerRuntime() {
			case "docker":
				binary = "docker"
			case "containerd", "crio":
				binary = "crictl"
			}

			rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "sudo", binary, "rmi", img))

			if err != nil {
				t.Errorf("failed to manually delete image %q : %v", rr.Command(), err)
			}
			// make sure the image is deleted.
			rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "sudo", "crictl", "inspecti", img))
			if err == nil {
				t.Errorf("expected an error  but got no error. image should not exist. ! cmd: %q", rr.Command())
			}
			// minikube cache reload.
			rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "cache", "reload"))
			if err != nil {
				t.Errorf("expected %q to run successfully but got error: %v", rr.Command(), err)
			}
			// make sure 'cache reload' brought back the manually deleted image.
			rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", "sudo", "crictl", "inspecti", img))
			if err != nil {
				t.Errorf("expected %q to run successfully but got error: %v", rr.Command(), err)
			}
		})

		// delete will clean up the cached images since they are global and all other tests will load it for no reason
		t.Run("delete", func(t *testing.T) {
			for _, img := range []string{"registry.k8s.io/pause:3.1", "registry.k8s.io/pause:latest"} {
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "cache", "delete", img))
				if err != nil {
					t.Errorf("failed to delete %s from cache. args %q: %v", img, rr.Command(), err)
				}
			}
		})
	})
}

// validateConfigCmd asserts basic "config" command functionality
func validateConfigCmd(ctx context.Context, t *testing.T, profile string) {
	// docs: Run `minikube config set/get/unset` to make sure configuration is modified correctly
	tests := []struct {
		args    []string
		wantOut string
		wantErr string
	}{
		{[]string{"unset", "cpus"}, "", ""},
		{[]string{"get", "cpus"}, "", "Error: specified key could not be found in config"},
		{[]string{"set", "cpus", "2"}, "", "! These changes will take effect upon a minikube delete and then a minikube start"},
		{[]string{"get", "cpus"}, "2", ""},
		{[]string{"unset", "cpus"}, "", ""},
		{[]string{"get", "cpus"}, "", "Error: specified key could not be found in config"},
	}

	for _, tc := range tests {
		args := append([]string{"-p", profile, "config"}, tc.args...)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
		if err != nil && tc.wantErr == "" {
			t.Errorf("failed to config minikube. args %q : %v", rr.Command(), err)
		}

		got := strings.TrimSpace(rr.Stdout.String())
		if got != tc.wantOut {
			t.Errorf("expected config output for %q to be -%q- but got *%q*", rr.Command(), tc.wantOut, got)
		}
		got = strings.TrimSpace(rr.Stderr.String())
		if got != tc.wantErr {
			t.Errorf("expected config error for %q to be -%q- but got *%q*", rr.Command(), tc.wantErr, got)
		}
	}
}

func checkSaneLogs(t *testing.T, logs string) {
	expectedWords := []string{"apiserver", "Linux", "kubelet", "Audit", "Last Start"}
	switch ContainerRuntime() {
	case "docker":
		expectedWords = append(expectedWords, "Docker")
	case "containerd":
		expectedWords = append(expectedWords, "containerd")
	case "crio":
		expectedWords = append(expectedWords, "crio")
	}

	for _, word := range expectedWords {
		if !strings.Contains(logs, word) {
			t.Errorf("expected minikube logs to include word: -%q- but got \n***%s***\n", word, logs)
		}
	}
}

// validateLogsCmd asserts basic "logs" command functionality
func validateLogsCmd(ctx context.Context, t *testing.T, profile string) {
	// docs: Run `minikube logs` and make sure the logs contains some keywords like `apiserver`, `Audit` and `Last Start`
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "logs"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	checkSaneLogs(t, rr.Stdout.String())
}

// validateLogsFileCmd asserts "logs --file" command functionality
func validateLogsFileCmd(ctx context.Context, t *testing.T, profile string) {
	dname := t.TempDir()
	logFileName := filepath.Join(dname, "logs.txt")

	// docs: Run `minikube logs --file logs.txt` to save the logs to a local file
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "logs", "--file", logFileName))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}

	logs, err := os.ReadFile(logFileName)
	if err != nil {
		t.Errorf("Failed to read logs output '%s': %v", logFileName, err)
	}

	// docs: Make sure the logs are correctly written
	checkSaneLogs(t, string(logs))
}

// validateProfileCmd asserts "profile" command functionality
func validateProfileCmd(ctx context.Context, t *testing.T, profile string) {
	t.Run("profile_not_create", func(t *testing.T) {
		// Profile command should not create a nonexistent profile
		nonexistentProfile := "lis"
		// docs: Run `minikube profile lis` and make sure the command doesn't fail for the non-existent profile `lis`
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", nonexistentProfile))
		if err != nil {
			t.Errorf("%s failed: %v", rr.Command(), err)
		}
		// docs: Run `minikube profile list --output json` to make sure the previous command doesn't create a new profile
		rr, err = Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "--output", "json"))
		if err != nil {
			t.Errorf("%s failed: %v", rr.Command(), err)
		}
		var profileJSON map[string][]map[string]interface{}
		err = json.Unmarshal(rr.Stdout.Bytes(), &profileJSON)
		if err != nil {
			t.Errorf("%s failed: %v", rr.Command(), err)
		}
		for profileK := range profileJSON {
			for _, p := range profileJSON[profileK] {
				var name = p["Name"]
				if name == nonexistentProfile {
					t.Errorf("minikube profile %s should not exist", nonexistentProfile)
				}
			}
		}
	})

	// docs: Run `minikube profile list` and make sure the profiles are correctly listed
	t.Run("profile_list", func(t *testing.T) {
		// helper function to run command then, return target profile line from table output.
		extractrofileListFunc := func(rr *RunResult) string {
			listLines := strings.Split(strings.TrimSpace(rr.Stdout.String()), "\n")
			for i := 3; i < (len(listLines) - 1); i++ {
				profileLine := listLines[i]
				if strings.Contains(profileLine, profile) {
					return profileLine
				}
			}
			return ""
		}

		// List profiles
		start := time.Now()
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list"))
		elapsed := time.Since(start)
		if err != nil {
			t.Errorf("failed to list profiles: args %q : %v", rr.Command(), err)
		}
		t.Logf("Took %q to run %q", elapsed, rr.Command())

		profileLine := extractrofileListFunc(rr)
		if profileLine == "" {
			t.Errorf("expected 'profile list' output to include %q but got *%q*. args: %q", profile, rr.Stdout.String(), rr.Command())
		}

		// List profiles with light option.
		start = time.Now()
		lrr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "-l"))
		lightElapsed := time.Since(start)
		if err != nil {
			t.Errorf("failed to list profiles: args %q : %v", lrr.Command(), err)
		}
		t.Logf("Took %q to run %q", lightElapsed, lrr.Command())

		profileLine = extractrofileListFunc(lrr)
		if profileLine == "" || !strings.Contains(profileLine, "Skipped") {
			t.Errorf("expected 'profile list' output to include %q with 'Skipped' status but got *%q*. args: %q", profile, rr.Stdout.String(), rr.Command())
		}

		if lightElapsed > 3*time.Second {
			t.Errorf("expected running time of '%q' is less than 3 seconds. Took %q ", lrr.Command(), lightElapsed)
		}
	})

	// docs: Run `minikube profile list -o JSON` and make sure the profiles are correctly listed as JSON output
	t.Run("profile_json_output", func(t *testing.T) {
		// helper function to run command then, return target profile object from json output.
		extractProfileObjFunc := func(rr *RunResult) *config.Profile {
			var jsonObject map[string][]config.Profile
			err := json.Unmarshal(rr.Stdout.Bytes(), &jsonObject)
			if err != nil {
				t.Errorf("failed to decode json from profile list: args %q: %v", rr.Command(), err)
				return nil
			}

			for _, profileObject := range jsonObject["valid"] {
				if profileObject.Name == profile {
					return &profileObject
				}
			}
			return nil
		}

		start := time.Now()
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "-o", "json"))
		elapsed := time.Since(start)
		if err != nil {
			t.Errorf("failed to list profiles with json format. args %q: %v", rr.Command(), err)
		}
		t.Logf("Took %q to run %q", elapsed, rr.Command())

		pr := extractProfileObjFunc(rr)
		if pr == nil {
			t.Errorf("expected the json of 'profile list' to include %q but got *%q*. args: %q", profile, rr.Stdout.String(), rr.Command())
		}

		start = time.Now()
		lrr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "-o", "json", "--light"))
		lightElapsed := time.Since(start)
		if err != nil {
			t.Errorf("failed to list profiles with json format. args %q: %v", lrr.Command(), err)
		}
		t.Logf("Took %q to run %q", lightElapsed, lrr.Command())

		pr = extractProfileObjFunc(lrr)
		if pr == nil || pr.Status != "Skipped" {
			t.Errorf("expected the json of 'profile list' to include 'Skipped' status for %q but got *%q*. args: %q", profile, lrr.Stdout.String(), lrr.Command())
		}

		if lightElapsed > 3*time.Second {
			t.Errorf("expected running time of '%q' is less than 3 seconds. Took %q ", lrr.Command(), lightElapsed)
		}
	})
}

// validateServiceCmd asserts basic "service" command functionality
func validateServiceCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	defer func() {
		if t.Failed() {
			t.Logf("service test failed - dumping debug information")
			t.Logf("-----------------------service failure post-mortem--------------------------------")
			ctx, cancel := context.WithTimeout(context.Background(), Minutes(2))
			defer cancel()
			rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "describe", "po", "hello-node"))
			if err != nil {
				t.Logf("%q failed: %v", rr.Command(), err)
			}
			t.Logf("hello-node pod describe:\n%s", rr.Stdout)

			rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "logs", "-l", "app=hello-node"))
			if err != nil {
				t.Logf("%q failed: %v", rr.Command(), err)
			}
			t.Logf("hello-node logs:\n%s", rr.Stdout)

			rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "describe", "svc", "hello-node"))
			if err != nil {
				t.Logf("%q failed: %v", rr.Command(), err)
			}
			t.Logf("hello-node svc describe:\n%s", rr.Stdout)
		}
	}()

	validateServiceCmdDeployApp(ctx, t, profile)
	validateServiceCmdList(ctx, t, profile)
	validateServiceCmdJSON(ctx, t, profile)
	validateServiceCmdHTTPS(ctx, t, profile)
	validateServiceCmdFormat(ctx, t, profile)
	validateServiceCmdURL(ctx, t, profile)
}

// validateServiceCmdDeployApp Create a new `registry.k8s.io/echoserver` deployment
func validateServiceCmdDeployApp(ctx context.Context, t *testing.T, profile string) {
	t.Run("DeployApp", func(t *testing.T) {
		var rr *RunResult
		var err error
		// registry.k8s.io/echoserver is not multi-arch
		if arm64Platform() {
			rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "deployment", "hello-node", "--image=registry.k8s.io/echoserver-arm:1.8"))
		} else {
			rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "deployment", "hello-node", "--image=registry.k8s.io/echoserver:1.8"))
		}

		if err != nil {
			t.Fatalf("failed to create hello-node deployment with this command %q: %v.", rr.Command(), err)
		}
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "expose", "deployment", "hello-node", "--type=NodePort", "--port=8080"))
		if err != nil {
			t.Fatalf("failed to expose hello-node deployment: %q : %v", rr.Command(), err)
		}

		if _, err := PodWait(ctx, t, profile, "default", "app=hello-node", Minutes(10)); err != nil {
			t.Fatalf("failed waiting for hello-node pod: %v", err)
		}
	})
}

// validateServiceCmdList Run `minikube service list` to make sure the newly created service is correctly listed in the output
func validateServiceCmdList(ctx context.Context, t *testing.T, profile string) {
	t.Run("List", func(t *testing.T) {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "service", "list"))
		if err != nil {
			t.Errorf("failed to do service list. args %q : %v", rr.Command(), err)
		}
		if !strings.Contains(rr.Stdout.String(), "hello-node") {
			t.Errorf("expected 'service list' to contain *hello-node* but got -%q-", rr.Stdout.String())
		}
	})
}

// validateServiceCmdJSON Run `minikube service list -o JSON` and make sure the services are correctly listed as JSON output
func validateServiceCmdJSON(ctx context.Context, t *testing.T, profile string) {
	t.Run("JSONOutput", func(t *testing.T) {
		targetSvcName := "hello-node"
		// helper function to run command then, return target service object from json output.
		extractServiceObjFunc := func(rr *RunResult) *service.SvcURL {
			var jsonObjects service.URLs
			if err := json.Unmarshal(rr.Stdout.Bytes(), &jsonObjects); err != nil {
				t.Fatalf("failed to decode json from profile list: args %q: %v", rr.Command(), err)
			}

			for _, svc := range jsonObjects {
				if svc.Name == targetSvcName {
					return &svc
				}
			}
			return nil
		}

		start := time.Now()
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "service", "list", "-o", "json"))
		if err != nil {
			t.Fatalf("failed to list services with json format. args %q: %v", rr.Command(), err)
		}
		elapsed := time.Since(start)
		t.Logf("Took %q to run %q", elapsed, rr.Command())

		pr := extractServiceObjFunc(rr)
		if pr == nil {
			t.Errorf("expected the json of 'service list' to include %q but got *%q*. args: %q", targetSvcName, rr.Stdout.String(), rr.Command())
		}
	})
}

// validateServiceCmdHTTPS Run `minikube service` with `--https --url` to make sure the HTTPS endpoint URL of the service is printed
func validateServiceCmdHTTPS(ctx context.Context, t *testing.T, profile string) {
	t.Run("HTTPS", func(t *testing.T) {
		cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(cmdCtx, Target(), "-p", profile, "service", "--namespace=default", "--https", "--url", "hello-node")
		rr, err := Run(t, cmd)
		if isUnexpectedServiceError(cmdCtx, err) {
			t.Fatalf("failed to get service url. args %q : %v", rr.Command(), err)
		}

		splits := strings.Split(rr.Stdout.String(), "|")
		var endpoint string
		// get the last endpoint in the output to test http to https
		for _, v := range splits {
			if strings.Contains(v, "http") {
				endpoint = strings.TrimSpace(v)
			}
		}
		t.Logf("found endpoint: %s", endpoint)

		u, err := url.Parse(endpoint)
		if err != nil {
			t.Fatalf("failed to parse service url endpoint %q: %v", endpoint, err)
		}
		if u.Scheme != "https" {
			t.Errorf("expected scheme for %s to be 'https' but got %q", endpoint, u.Scheme)
		}
	})
}

// validateServiceCmdFormat Run `minikube service` with `--url --format={{.IP}}` to make sure the IP address of the service is printed
func validateServiceCmdFormat(ctx context.Context, t *testing.T, profile string) {
	t.Run("Format", func(t *testing.T) {
		cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(cmdCtx, Target(), "-p", profile, "service", "hello-node", "--url", "--format={{.IP}}")
		rr, err := Run(t, cmd)
		if isUnexpectedServiceError(cmdCtx, err) {
			t.Errorf("failed to get service url with custom format. args %q: %v", rr.Command(), err)
		}

		stringIP := strings.TrimSpace(rr.Stdout.String())

		if ip := net.ParseIP(stringIP); ip == nil {
			t.Fatalf("%q is not a valid IP", stringIP)
		}
	})
}

// validateServiceCmdURL Run `minikube service` with a regular `--url` to make sure the HTTP endpoint URL of the service is printed
func validateServiceCmdURL(ctx context.Context, t *testing.T, profile string) {
	t.Run("URL", func(t *testing.T) {
		cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(cmdCtx, Target(), "-p", profile, "service", "hello-node", "--url")
		rr, err := Run(t, cmd)
		if isUnexpectedServiceError(cmdCtx, err) {
			t.Errorf("failed to get service url. args: %q: %v", rr.Command(), err)
		}

		endpoint := strings.TrimSpace(rr.Stdout.String())
		t.Logf("found endpoint for hello-node: %s", endpoint)

		u, err := url.Parse(endpoint)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", endpoint, err)
		}

		if u.Scheme != "http" {
			t.Fatalf("expected scheme to be -%q- got scheme: *%q*", "http", u.Scheme)
		}
	})
}

// isUnexpectedServiceError is used to prevent failing ServiceCmd tests on Docker Desktop due to DeadlineExceeded errors.
// Due to networking constraints Docker Desktop requires creating an SSH tunnel to connect to a service. This command has
// to be left running to keep the SSH tunnel connected, so for the ServiceCmd tests we set a timeout context so we can
// check the output and then the command is terminated, otherwise it would keep runnning forever. So if using Docker
// Desktop and the DeadlineExceeded, consider it an expected error.
func isUnexpectedServiceError(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if !NeedsPortForward() {
		return true
	}
	return ctx.Err() != context.DeadlineExceeded
}

func validateServiceCmdConnect(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	defer func() {
		if t.Failed() {
			t.Logf("service test failed - dumping debug information")
			t.Logf("-----------------------service failure post-mortem--------------------------------")
			ctx, cancel := context.WithTimeout(context.Background(), Minutes(2))
			defer cancel()
			rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "describe", "po", "hello-node-connect"))
			if err != nil {
				t.Logf("%q failed: %v", rr.Command(), err)
			}
			t.Logf("hello-node pod describe:\n%s", rr.Stdout)

			rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "logs", "-l", "app=hello-node-connect"))
			if err != nil {
				t.Logf("%q failed: %v", rr.Command(), err)
			}
			t.Logf("hello-node logs:\n%s", rr.Stdout)

			rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "describe", "svc", "hello-node-connect"))
			if err != nil {
				t.Logf("%q failed: %v", rr.Command(), err)
			}
			t.Logf("hello-node svc describe:\n%s", rr.Stdout)
		}
	}()

	var rr *RunResult
	var err error
	// docs: Create a new `registry.k8s.io/echoserver` deployment
	// registry.k8s.io/echoserver is not multi-arch
	if arm64Platform() {
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "deployment", "hello-node-connect", "--image=registry.k8s.io/echoserver-arm:1.8"))
	} else {
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "deployment", "hello-node-connect", "--image=registry.k8s.io/echoserver:1.8"))
	}

	if err != nil {
		t.Fatalf("failed to create hello-node deployment with this command %q: %v.", rr.Command(), err)
	}
	rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "expose", "deployment", "hello-node-connect", "--type=NodePort", "--port=8080"))
	if err != nil {
		t.Fatalf("failed to expose hello-node deployment: %q : %v", rr.Command(), err)
	}

	if _, err := PodWait(ctx, t, profile, "default", "app=hello-node-connect", Minutes(10)); err != nil {
		t.Fatalf("failed waiting for hello-node pod: %v", err)
	}

	cmdContext := exec.CommandContext(ctx, Target(), "-p", profile, "service", "hello-node-connect", "--url")
	if NeedsPortForward() {
		t.Skipf("test is broken for port-forwarded drivers: https://github.com/kubernetes/minikube/issues/7383")
	}
	// docs: Run `minikube service` with a regular `--url` to make sure the HTTP endpoint URL of the service is printed
	rr, err = Run(t, cmdContext)
	if err != nil {
		t.Errorf("failed to get service url. args: %q: %v", rr.Command(), err)
	}

	endpoint := strings.TrimSpace(rr.Stdout.String())
	t.Logf("found endpoint for hello-node-connect: %s", endpoint)

	// docs: Make sure we can hit the endpoint URL with an HTTP GET request
	fetch := func() error {
		resp, err := http.Get(endpoint)
		if err != nil {
			t.Logf("error fetching %s: %v", endpoint, err)
			return err
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("error reading body from %s: %v", endpoint, err)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			t.Logf("%s: unexpected status code %d - body:\n%s", endpoint, resp.StatusCode, body)
		} else {
			t.Logf("%s: success! body:\n%s", endpoint, body)
		}
		return nil
	}

	if err = retry.Expo(fetch, 1*time.Second, Seconds(30)); err != nil {
		t.Errorf("failed to fetch %s: %v", endpoint, err)
	}
}

// validateAddonsCmd asserts basic "addon" command functionality
func validateAddonsCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs: Run `minikube addons list` to list the addons in a tabular format
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "list"))
	if err != nil {
		t.Errorf("failed to do addon list: args %q : %v", rr.Command(), err)
	}
	// docs: Make sure `dashboard`, `ingress` and `ingress-dns` is listed as available addons
	for _, a := range []string{"dashboard", "ingress", "ingress-dns"} {
		if !strings.Contains(rr.Output(), a) {
			t.Errorf("expected 'addon list' output to include -%q- but got *%s*", a, rr.Output())
		}
	}

	// docs: Run `minikube addons list -o JSON` lists the addons in JSON format
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "list", "-o", "json"))
	if err != nil {
		t.Errorf("failed to do addon list with json output. args %q: %v", rr.Command(), err)
	}
	var jsonObject map[string]interface{}
	err = json.Unmarshal(rr.Stdout.Bytes(), &jsonObject)
	if err != nil {
		t.Errorf("failed to decode addon list json output : %v", err)
	}
}

// validateSSHCmd asserts basic "ssh" command functionality
func validateSSHCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	if NoneDriver() {
		t.Skipf("skipping: ssh unsupported by none")
	}
	mctx, cancel := context.WithTimeout(ctx, Minutes(1))
	defer cancel()

	want := "hello"

	// docs: Run `minikube ssh echo hello` to make sure we can SSH into the minikube container and run an command
	rr, err := Run(t, exec.CommandContext(mctx, Target(), "-p", profile, "ssh", "echo hello"))
	if mctx.Err() == context.DeadlineExceeded {
		t.Errorf("failed to run command by deadline. exceeded timeout : %s", rr.Command())
	}
	if err != nil {
		t.Errorf("failed to run an ssh command. args %q : %v", rr.Command(), err)
	}
	// trailing whitespace differs between native and external SSH clients, so let's trim it and call it a day
	if strings.TrimSpace(rr.Stdout.String()) != want {
		t.Errorf("expected minikube ssh command output to be -%q- but got *%q*. args %q", want, rr.Stdout.String(), rr.Command())
	}

	// docs: Run `minikube ssh cat /etc/hostname` as well to make sure the command is run inside minikube
	// because it  is not clear if echo was run inside minikube on the powershell
	// so better to test something inside minikube, that is meaningful per profile
	// in this case /etc/hostname is same as the profile name
	want = profile
	rr, err = Run(t, exec.CommandContext(mctx, Target(), "-p", profile, "ssh", "cat /etc/hostname"))
	if mctx.Err() == context.DeadlineExceeded {
		t.Errorf("failed to run command by deadline. exceeded timeout : %s", rr.Command())
	}

	if err != nil {
		t.Errorf("failed to run an ssh command. args %q : %v", rr.Command(), err)
	}
	// trailing whitespace differs between native and external SSH clients, so let's trim it and call it a day
	if strings.TrimSpace(rr.Stdout.String()) != want {
		t.Errorf("expected minikube ssh command output to be -%q- but got *%q*. args %q", want, rr.Stdout.String(), rr.Command())
	}
}

// validateCpCmd asserts basic "cp" command functionality
func validateCpCmd(ctx context.Context, t *testing.T, profile string) {
	// docs(skip): Skips `none` driver since `cp` is not supported
	if NoneDriver() {
		t.Skipf("skipping: cp is unsupported by none driver")
	}

	// docs: Run `minikube cp ...` to copy a file to the minikube node
	// docs: Run `minikube ssh sudo cat ...` to print out the copied file within minikube
	// docs: make sure the file is correctly copied

	srcPath := cpTestLocalPath()
	dstPath := cpTestMinikubePath()

	// copy to node
	testCpCmd(ctx, t, profile, "", srcPath, "", dstPath)

	// copy from node
	tmpDir := t.TempDir()

	tmpPath := filepath.Join(tmpDir, "cp-test.txt")
	testCpCmd(ctx, t, profile, profile, dstPath, "", tmpPath)

	// copy to nonexistent directory structure
	testCpCmd(ctx, t, profile, "", srcPath, "", "/tmp/does/not/exist/cp-test.txt")
}

// validateMySQL validates a minimalist MySQL deployment
func validateMySQL(ctx context.Context, t *testing.T, profile string) {
	// docs(skip): Skips for ARM64 architecture since it's not supported by MySQL
	if arm64Platform() {
		t.Skip("arm64 is not supported by mysql. Skip the test. See https://github.com/kubernetes/minikube/issues/10144")
	}

	defer PostMortemLogs(t, profile)

	// docs: Run `kubectl replace --force -f testdata/mysql/yaml`
	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "mysql.yaml")))
	if err != nil {
		t.Fatalf("failed to kubectl replace mysql: args %q failed: %v", rr.Command(), err)
	}

	// docs: Wait for the `mysql` pod to be running
	names, err := PodWait(ctx, t, profile, "default", "app=mysql", Minutes(10))
	if err != nil {
		t.Fatalf("failed waiting for mysql pod: %v", err)
	}

	// docs: Run `mysql -e show databases;` inside the MySQL pod to verify MySQL is up and running
	// docs: Retry with exponential backoff if failed, as `mysqld` first comes up without users configured. Scan for names in case of a reschedule.
	mysql := func() error {
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "exec", names[0], "--", "mysql", "-ppassword", "-e", "show databases;"))
		return err
	}
	if err = retry.Expo(mysql, 1*time.Second, Minutes(5)); err != nil {
		t.Errorf("failed to exec 'mysql -ppassword -e show databases;': %v", err)
	}
}

// vmSyncTestPath is where the test file will be synced into the VM
func vmSyncTestPath() string {
	return fmt.Sprintf("/etc/test/nested/copy/%d/hosts", os.Getpid())
}

// localSyncTestPath is where the test file will be synced into the VM
func localSyncTestPath() string {
	return filepath.Join(localpath.MiniPath(), "/files", vmSyncTestPath())
}

// testCert is name of the test certificate installed
func testCert() string {
	return fmt.Sprintf("%d.pem", os.Getpid())
}

// testFileCert is name of the test certificate installed
func testFileCert() string {
	return fmt.Sprintf("%d2.pem", os.Getpid())
}

// localTestCertPath is where certs can be synced from the local host into the VM
// precisely, it's $MINIKUBE_HOME/certs
func localTestCertPath() string {
	return filepath.Join(localpath.MiniPath(), "/certs", testCert())
}

// localTestCertFilesPath is an alternate location where certs can be synced into the minikube VM
// precisely, it's $MINIKUBE_HOME/files/etc/ssl/certs
func localTestCertFilesPath() string {
	return filepath.Join(localpath.MiniPath(), "/files/etc/ssl/certs", testFileCert())
}

// localEmptyCertPath is where the test file will be synced into the VM
func localEmptyCertPath() string {
	return filepath.Join(localpath.MiniPath(), "/certs", fmt.Sprintf("%d_empty.pem", os.Getpid()))
}

// Copy extra file into minikube home folder for file sync test
func setupFileSync(_ context.Context, t *testing.T, _ string) {
	p := localSyncTestPath()
	t.Logf("local sync path: %s", p)
	syncFile := filepath.Join(*testdataDir, "sync.test")
	err := cp.Copy(syncFile, p)
	if err != nil {
		t.Fatalf("failed to copy testdata/sync.test: %v", err)
	}

	testPem := filepath.Join(*testdataDir, "minikube_test.pem")

	// Write to a temp file for an atomic write
	tmpPem := localTestCertPath() + ".pem"
	if err := cp.Copy(testPem, tmpPem); err != nil {
		t.Fatalf("failed to copy %s: %v", testPem, err)
	}

	if err := os.Rename(tmpPem, localTestCertPath()); err != nil {
		t.Fatalf("failed to rename %s: %v", tmpPem, err)
	}

	want, err := os.Stat(testPem)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	got, err := os.Stat(localTestCertPath())
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	if want.Size() != got.Size() {
		t.Errorf("%s size=%d, want %d", localTestCertPath(), got.Size(), want.Size())
	}

	testPem2 := filepath.Join(*testdataDir, "minikube_test2.pem")
	tmpPem2 := localTestCertFilesPath() + ".pem"
	if err := cp.Copy(testPem2, tmpPem2); err != nil {
		t.Fatalf("failed to copy %s: %v", testPem2, err)
	}

	if err := os.Rename(tmpPem2, localTestCertFilesPath()); err != nil {
		t.Fatalf("failed to rename %s: %v", tmpPem2, err)
	}

	want, err = os.Stat(testPem2)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	got, err = os.Stat(localTestCertFilesPath())
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	if want.Size() != got.Size() {
		t.Errorf("%s size=%d, want %d", localTestCertFilesPath(), got.Size(), want.Size())
	}

	// Create an empty file just to mess with people
	if _, err := os.Create(localEmptyCertPath()); err != nil {
		t.Fatalf("create failed: %v", err)
	}
}

// validateFileSync to check existence of the test file
func validateFileSync(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// docs(skip): Skips on `none` driver since SSH is not supported
	if NoneDriver() {
		t.Skipf("skipping: ssh unsupported by none")
	}

	// docs: Test files have been synced into minikube in the previous step `setupFileSync`
	vp := vmSyncTestPath()
	t.Logf("Checking for existence of %s within VM", vp)
	// docs: Check the existence of the test file
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("sudo cat %s", vp)))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}
	got := rr.Stdout.String()
	t.Logf("file sync test content: %s", got)

	syncFile := filepath.Join(*testdataDir, "sync.test")
	expected, err := os.ReadFile(syncFile)
	if err != nil {
		t.Errorf("failed to read test file 'testdata/sync.test' : %v", err)
	}

	// docs: Make sure the file is correctly synced
	if diff := cmp.Diff(string(expected), got); diff != "" {
		t.Errorf("/etc/sync.test content mismatch (-want +got):\n%s", diff)
	}
}

// validateCertSync checks to make sure a custom cert has been copied into the minikube guest and installed correctly
func validateCertSync(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	if NoneDriver() {
		t.Skipf("skipping: ssh unsupported by none")
	}

	testPem := filepath.Join(*testdataDir, "minikube_test.pem")
	want, err := os.ReadFile(testPem)
	if err != nil {
		t.Errorf("test file not found: %v", err)
	}

	// docs: Check both the installed & reference certs and make sure they are symlinked
	paths := []string{
		path.Join("/etc/ssl/certs", testCert()),
		path.Join("/usr/share/ca-certificates", testCert()),
		// hashed path generated by: 'openssl x509 -hash -noout -in testCert()'
		"/etc/ssl/certs/51391683.0",
	}
	for _, vp := range paths {
		t.Logf("Checking for existence of %s within VM", vp)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("sudo cat %s", vp)))
		if err != nil {
			t.Errorf("failed to check existence of %q inside minikube. args %q: %v", vp, rr.Command(), err)
		}

		// Strip carriage returned by ssh
		got := strings.ReplaceAll(rr.Stdout.String(), "\r", "")
		if diff := cmp.Diff(string(want), got); diff != "" {
			t.Errorf("failed verify pem file. minikube_test.pem -> %s mismatch (-want +got):\n%s", vp, diff)
		}
	}

	testPem2 := filepath.Join(*testdataDir, "minikube_test2.pem")
	want, err = os.ReadFile(testPem2)
	if err != nil {
		t.Errorf("test file not found: %v", err)
	}

	// Check both the installed & reference certs (they should be symlinked)
	paths = []string{
		path.Join("/etc/ssl/certs", testFileCert()),
		path.Join("/usr/share/ca-certificates", testFileCert()),
		// hashed path generated by: 'openssl x509 -hash -noout -in testCert()'
		"/etc/ssl/certs/3ec20f2e.0",
	}
	for _, vp := range paths {
		t.Logf("Checking for existence of %s within VM", vp)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("sudo cat %s", vp)))
		if err != nil {
			t.Errorf("failed to check existence of %q inside minikube. args %q: %v", vp, rr.Command(), err)
		}

		// Strip carriage returned by ssh
		got := strings.ReplaceAll(rr.Stdout.String(), "\r", "")
		if diff := cmp.Diff(string(want), got); diff != "" {
			t.Errorf("failed verify pem file. minikube_test2.pem -> %s mismatch (-want +got):\n%s", vp, diff)
		}
	}
}

// validateNotActiveRuntimeDisabled asserts that for a given runtime, the other runtimes are disabled, for example for `containerd` runtime, `docker` and `crio` needs to be not running
func validateNotActiveRuntimeDisabled(ctx context.Context, t *testing.T, profile string) {
	if NoneDriver() {
		t.Skip("skipping on none driver, minikube does not control the runtime of user on the none driver.")
	}
	disableMap := map[string][]string{
		"docker":     {"crio"},
		"containerd": {"docker", "crio"},
		"crio":       {"docker", "containerd"},
	}

	expectDisable := disableMap[ContainerRuntime()]
	for _, cr := range expectDisable {
		// docs: For each container runtime, run `minikube ssh sudo systemctl is-active ...` and make sure the other container runtimes are not running
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("sudo systemctl is-active %s", cr)))
		got := rr.Stdout.String()
		if err != nil && !strings.Contains(got, "inactive") {
			t.Logf("output of %s: %v", rr.Output(), err)
		}
		if !strings.Contains(got, "inactive") {
			t.Errorf("For runtime %q: expected %q to be inactive but got %q ", ContainerRuntime(), cr, got)
		}

	}
}

// validateUpdateContextCmd asserts basic "update-context" command functionality
func validateUpdateContextCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	tests := []struct {
		name       string
		kubeconfig []byte
		want       []byte
	}{
		{
			name:       "no changes",
			kubeconfig: nil,
			want:       []byte("No changes"),
		},
		{
			name: "no minikube cluster",
			kubeconfig: []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/la-croix/apiserver.crt
    server: 192.168.1.1:8080
  name: la-croix
contexts:
- context:
    cluster: la-croix
    user: la-croix
  name: la-croix
current-context: la-croix
kind: Config
preferences: {}
users:
- name: la-croix
  user:
    client-certificate: /home/la-croix/apiserver.crt
    client-key: /home/la-croix/apiserver.key
`),
			want: []byte("context has been updated"),
		},
		{
			name: "no clusters",
			kubeconfig: []byte(`
apiVersion: v1
clusters:
contexts:
kind: Config
preferences: {}
users:
`),
			want: []byte("context has been updated"),
		},
	}

	for _, tc := range tests {
		tc := tc

		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("Unable to run more tests (deadline exceeded)")
		}

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := exec.CommandContext(ctx, Target(), "-p", profile, "update-context", "--alsologtostderr", "-v=2")
			if tc.kubeconfig != nil {
				tf, err := os.CreateTemp("", "kubeconfig")
				if err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(tf.Name(), tc.kubeconfig, 0644); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					os.Remove(tf.Name())
				})

				c.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", tf.Name()))
			}

			// docs: Run `minikube update-context`
			rr, err := Run(t, c)
			if err != nil {
				t.Errorf("failed to run minikube update-context: args %q: %v", rr.Command(), err)
			}

			// docs: Make sure the context has been correctly updated by checking the command output
			if !bytes.Contains(rr.Stdout.Bytes(), tc.want) {
				t.Errorf("update-context: got=%q, want=*%q*", rr.Stdout.Bytes(), tc.want)
			}
		})
	}
}

// startProxyWithCustomCerts mimics starts a proxy with custom certs by using mitmproxy and installing its certs
func startProxyWithCustomCerts(ctx context.Context, t *testing.T) error {
	// Download the mitmproxy bundle for mitmdump
	_, err := Run(t, exec.CommandContext(ctx, "curl", "-LO", "https://snapshots.mitmproxy.org/6.0.2/mitmproxy-6.0.2-linux.tar.gz"))
	if err != nil {
		return errors.Wrap(err, "download mitmproxy tar")
	}
	defer func() {
		err := os.Remove("mitmproxy-6.0.2-linux.tar.gz")
		if err != nil {
			t.Logf("remove tarball: %v", err)
		}
	}()

	mitmDir := t.TempDir()

	_, err = Run(t, exec.CommandContext(ctx, "tar", "xzf", "mitmproxy-6.0.2-linux.tar.gz", "-C", mitmDir))
	if err != nil {
		return errors.Wrap(err, "untar mitmproxy tar")
	}

	// Start mitmdump in the background, this will create the needed certs
	// and provide the necessary proxy at 127.0.0.1:8080
	mitmRR, err := Start(t, exec.CommandContext(ctx, path.Join(mitmDir, "mitmdump"), "--set", fmt.Sprintf("confdir=%s", mitmDir)))
	if err != nil {
		return errors.Wrap(err, "starting mitmproxy")
	}

	// Store it for cleanup later
	mitm = mitmRR

	// Add a symlink from the cert to the correct directory
	certFile := path.Join(mitmDir, "mitmproxy-ca-cert.pem")
	// wait 15 seconds for the certs to show up
	_, err = os.Stat(certFile)
	tries := 1
	for os.IsNotExist(err) {
		time.Sleep(1 * time.Second)
		tries++
		if tries > 15 {
			break
		}
		_, err = os.Stat(certFile)
	}
	if os.IsNotExist(err) {
		return errors.Wrap(err, "cert files never showed up")
	}

	destCertPath := path.Join("/etc/ssl/certs", "mitmproxy-ca-cert.pem")
	symLinkCmd := fmt.Sprintf("ln -fs %s %s", certFile, destCertPath)
	if _, err := Run(t, exec.CommandContext(ctx, "sudo", "/bin/bash", "-c", symLinkCmd)); err != nil {
		return errors.Wrap(err, "cert symlink")
	}

	// Add a symlink of the form {hash}.0
	rr, err := Run(t, exec.CommandContext(ctx, "openssl", "x509", "-hash", "-noout", "-in", certFile))
	if err != nil {
		return errors.Wrap(err, "cert hashing")
	}
	stringHash := strings.TrimSpace(rr.Stdout.String())
	hashLink := path.Join("/etc/ssl/certs", fmt.Sprintf("%s.0", stringHash))

	hashCmd := fmt.Sprintf("test -L %s || ln -fs %s %s", hashLink, destCertPath, hashLink)
	if _, err := Run(t, exec.CommandContext(ctx, "sudo", "/bin/bash", "-c", hashCmd)); err != nil {
		return errors.Wrap(err, "cert hash symlink")
	}

	return nil
}

// startHTTPProxy runs a local http proxy and sets the env vars for it.
func startHTTPProxy(t *testing.T) (*http.Server, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get an open port")
	}

	addr := fmt.Sprintf("localhost:%d", port)
	proxy := goproxy.NewProxyHttpServer()
	srv := &http.Server{Addr: addr, Handler: proxy}
	go func(s *http.Server, t *testing.T) {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			t.Errorf("Failed to start http server for proxy mock")
		}
	}(srv, t)
	return srv, nil
}

func startMinikubeWithProxy(ctx context.Context, t *testing.T, profile string, proxyEnv string, addr string) {
	// Use more memory so that we may reliably fit MySQL and nginx
	memoryFlag := "--memory=4000"
	// to avoid failure for mysq/pv on virtualbox on darwin on free github actions,
	if detect.GithubActionRunner() && VirtualboxDriver() {
		memoryFlag = "--memory=6000"
	}
	// passing --api-server-port so later verify it didn't change in soft start.
	startArgs := append([]string{"start", "-p", profile, memoryFlag, fmt.Sprintf("--apiserver-port=%d", apiPortTest), "--wait=all"}, StartArgsWithContext(ctx)...)
	c := exec.CommandContext(ctx, Target(), startArgs...)
	env := os.Environ()
	env = append(env, fmt.Sprintf("%s=%s", proxyEnv, addr))
	env = append(env, "NO_PROXY=")
	c.Env = env
	rr, err := Run(t, c)
	if err != nil {
		t.Errorf("failed minikube start. args %q: %v", rr.Command(), err)
	}

	want := "Found network options:"
	if !strings.Contains(rr.Stdout.String(), want) {
		t.Errorf("start stdout=%s, want: *%s*", rr.Stdout.String(), want)
	}

	want = "You appear to be using a proxy"
	if !strings.Contains(rr.Stderr.String(), want) {
		t.Errorf("start stderr=%s, want: *%s*", rr.Stderr.String(), want)
	}
}

// validateVersionCmd asserts `minikube version` command works fine for both --short and --components
func validateVersionCmd(ctx context.Context, t *testing.T, profile string) {

	// docs: Run `minikube version --short` and make sure the returned version is a valid semver
	t.Run("short", func(t *testing.T) {
		MaybeParallel(t)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "version", "--short"))
		if err != nil {
			t.Errorf("failed to get version --short: %v", err)
		}

		_, err = semver.Make(strings.TrimSpace(strings.Trim(rr.Stdout.String(), "v")))
		if err != nil {
			t.Errorf("failed to get a valid semver for minikube version --short:%s %v", rr.Output(), err)
		}
	})

	// docs: Run `minikube version --components` and make sure the component versions are returned
	t.Run("components", func(t *testing.T) {
		MaybeParallel(t)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "version", "-o=json", "--components"))
		if err != nil {
			t.Errorf("error version: %v", err)
		}
		got := rr.Stdout.String()
		for _, c := range []string{"buildctl", "commit", "containerd", "crictl", "crio", "ctr", "docker", "minikubeVersion", "podman", "run", "crun"} {
			if !strings.Contains(got, c) {
				t.Errorf("expected to see %q in the minikube version --components but got:\n%s", c, got)
			}

		}
	})

}

// validateLicenseCmd asserts that the `minikube license` command downloads and untars the licenses
// Note: This test will fail on release PRs as the licenses file for the new version won't be uploaded at that point
func validateLicenseCmd(ctx context.Context, t *testing.T, _ string) {
	if rr, err := Run(t, exec.CommandContext(ctx, Target(), "license")); err != nil {
		t.Fatalf("command %q failed: %v", rr.Stdout.String(), err)
	}
	defer os.Remove("./licenses")
	files, err := os.ReadDir("./licenses")
	if err != nil {
		t.Fatalf("failed to read licenses dir: %v", err)
	}
	expectedDir := "cloud.google.com"
	found := false
	for _, file := range files {
		if file.Name() == expectedDir {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected licenses dir to contain %s dir, but was not found", expectedDir)
	}
	data, err := os.ReadFile("./licenses/cloud.google.com/go/compute/metadata/LICENSE")
	if err != nil {
		t.Fatalf("failed to read license: %v", err)
	}
	expectedString := "Apache License"
	if !strings.Contains(string(data), expectedString) {
		t.Errorf("expected license file to contain %q, but was not found", expectedString)
	}
}

// validateInvalidService makes sure minikube will not start a tunnel for an unavailable service that has no running pods
func validateInvalidService(ctx context.Context, t *testing.T, profile string) {

	// try to start an invalid service. This service is linked to a pod whose image name is invalid, so this pod will never become running
	rrApply, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "apply", "-f", filepath.Join(*testdataDir, "invalidsvc.yaml")))
	if err != nil {
		t.Fatalf("%s failed: %v", rrApply.Command(), err)
	}
	defer func() {
		// Cleanup test configurations in advance of future tests
		rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "delete", "-f", filepath.Join(*testdataDir, "invalidsvc.yaml")))
		if err != nil {
			t.Fatalf("clean up %s failed: %v", rr.Command(), err)
		}
	}()
	time.Sleep(3 * time.Second)

	// try to expose a service, this action is supposed to fail
	rrService, err := Run(t, exec.CommandContext(ctx, Target(), "service", "invalid-svc", "-p", profile))
	if err == nil || rrService.ExitCode == 0 {
		t.Fatalf("%s should have failed: ", rrService.Command())
	}
}
