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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/util/retry"

	"github.com/elazarl/goproxy"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/otiai10/copy"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"golang.org/x/build/kubernetes/api"
)

// validateFunc are for subtests that share a single setup
type validateFunc func(context.Context, *testing.T, string)

// used in validateStartWithProxy and validateSoftStart
var apiPortTest = 8441

// TestFunctional are functionality tests which can safely share a profile in parallel
func TestFunctional(t *testing.T) {

	profile := UniqueProfileName("functional")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
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
			{"SoftStart", validateSoftStart},                // do a soft start. ensure config didnt change.
			{"KubeContext", validateKubeContext},            // Racy: must come immediately after "minikube start"
			{"KubectlGetPods", validateKubectlGetPods},      // Make sure apiserver is up
			{"CacheCmd", validateCacheCmd},                  // Caches images needed for subsequent tests because of proxy
			{"MinikubeKubectlCmd", validateMinikubeKubectl}, // Make sure `minikube kubectl` works
			{"MinikubeKubectlCmdDirectly", validateMinikubeKubectlDirectCall},
			{"ExtraConfig", validateExtraConfig}, // Ensure extra cmdline config change is saved
		}
		for _, tc := range tests {
			tc := tc
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Unable to run more tests (deadline exceeded)")
			}
			t.Run(tc.name, func(t *testing.T) {
				tc.validator(ctx, t, profile)
			})
		}
	})

	// Parallelized tests
	t.Run("parallel", func(t *testing.T) {
		tests := []struct {
			name      string
			validator validateFunc
		}{
			{"ComponentHealth", validateComponentHealth},
			{"ConfigCmd", validateConfigCmd},
			{"DashboardCmd", validateDashboardCmd},
			{"DryRun", validateDryRun},
			{"StatusCmd", validateStatusCmd},
			{"LogsCmd", validateLogsCmd},
			{"MountCmd", validateMountCmd},
			{"ProfileCmd", validateProfileCmd},
			{"ServiceCmd", validateServiceCmd},
			{"AddonsCmd", validateAddonsCmd},
			{"PersistentVolumeClaim", validatePersistentVolumeClaim},
			{"TunnelCmd", validateTunnelCmd},
			{"SSHCmd", validateSSHCmd},
			{"MySQL", validateMySQL},
			{"FileSync", validateFileSync},
			{"CertSync", validateCertSync},
			{"UpdateContextCmd", validateUpdateContextCmd},
			{"DockerEnv", validateDockerEnv},
			{"NodeLabels", validateNodeLabels},
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

// validateNodeLabels checks if minikube cluster is created with correct kubernetes's node label
func validateNodeLabels(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "nodes", "--output=go-template", "--template='{{range $k, $v := (index .items 0).metadata.labels}}{{$k}} {{end}}'"))
	if err != nil {
		t.Errorf("failed to 'kubectl get nodes' with args %q: %v", rr.Command(), err)
	}
	expectedLabels := []string{"minikube.k8s.io/commit", "minikube.k8s.io/version", "minikube.k8s.io/updated_at", "minikube.k8s.io/name"}
	for _, el := range expectedLabels {
		if !strings.Contains(rr.Output(), el) {
			t.Errorf("expected to have label %q in node labels but got : %s", el, rr.Output())
		}
	}
}

// check functionality of minikube after evaling docker-env
// TODO: Add validatePodmanEnv for crio runtime: #10231
func validateDockerEnv(ctx context.Context, t *testing.T, profile string) {
	if cr := ContainerRuntime(); cr != "docker" {
		t.Skipf("only validate docker env with docker container runtime, currently testing %s", cr)
	}
	defer PostMortemLogs(t, profile)
	mctx, cancel := context.WithTimeout(ctx, Seconds(120))
	defer cancel()
	var rr *RunResult
	var err error
	if runtime.GOOS == "windows" {
		c := exec.CommandContext(mctx, "powershell.exe", "-NoProfile", "-NonInteractive", Target()+" -p "+profile+" docker-env | Invoke-Expression ;"+Target()+" status -p "+profile)
		rr, err = Run(t, c)
	} else {
		c := exec.CommandContext(mctx, "/bin/bash", "-c", "eval $("+Target()+" -p "+profile+" docker-env) && "+Target()+" status -p "+profile)
		// we should be able to get minikube status with a bash which evaled docker-env
		rr, err = Run(t, c)
	}
	if mctx.Err() == context.DeadlineExceeded {
		t.Errorf("failed to run the command by deadline. exceeded timeout. %s", rr.Command())
	}
	if err != nil {
		t.Fatalf("failed to do status after eval-ing docker-env. error: %v", err)
	}
	if !strings.Contains(rr.Output(), "Running") {
		t.Fatalf("expected status output to include 'Running' after eval docker-env but got: *%s*", rr.Output())
	}

	mctx, cancel = context.WithTimeout(ctx, Seconds(60))
	defer cancel()
	// do a eval $(minikube -p profile docker-env) and check if we are point to docker inside minikube
	if runtime.GOOS == "windows" { // testing docker-env eval in powershell
		c := exec.CommandContext(mctx, "powershell.exe", "-NoProfile", "-NonInteractive", Target(), "-p "+profile+" docker-env | Invoke-Expression ; docker images")
		rr, err = Run(t, c)
	} else {
		c := exec.CommandContext(mctx, "/bin/bash", "-c", "eval $("+Target()+" -p "+profile+" docker-env) && docker images")
		rr, err = Run(t, c)
	}

	if mctx.Err() == context.DeadlineExceeded {
		t.Errorf("failed to run the command in 30 seconds. exceeded 30s timeout. %s", rr.Command())
	}

	if err != nil {
		t.Fatalf("failed to run minikube docker-env. args %q : %v ", rr.Command(), err)
	}

	expectedImgInside := "gcr.io/k8s-minikube/storage-provisioner"
	if !strings.Contains(rr.Output(), expectedImgInside) {
		t.Fatalf("expected 'docker images' to have %q inside minikube. but the output is: *%s*", expectedImgInside, rr.Output())
	}

}

func validateStartWithProxy(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	srv, err := startHTTPProxy(t)
	if err != nil {
		t.Fatalf("failed to set up the test proxy: %s", err)
	}

	// Use more memory so that we may reliably fit MySQL and nginx
	// changing api server so later in soft start we verify it didn't change
	startArgs := append([]string{"start", "-p", profile, "--memory=4000", "--cpus=4", fmt.Sprintf("--apiserver-port=%d", apiPortTest), "--wait=true"}, StartArgs()...)
	c := exec.CommandContext(ctx, Target(), startArgs...)
	env := os.Environ()
	env = append(env, fmt.Sprintf("HTTP_PROXY=%s", srv.Addr))
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

	t.Run("Audit", func(t *testing.T) {
		got, err := auditContains(profile)
		if err != nil {
			t.Fatalf("failed to check audit log: %v", err)
		}
		if !got {
			t.Errorf("audit.json does not contain the profile %q", profile)
		}
	})
}

// validateSoftStart validates that after minikube already started, a "minikube start" should not change the configs.
func validateSoftStart(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	start := time.Now()
	// the test before this had been start with --apiserver-port=8441
	beforeCfg, err := config.LoadProfile(profile)
	if err != nil {
		t.Fatalf("error reading cluster config before soft start: %v", err)
	}
	if beforeCfg.Config.KubernetesConfig.NodePort != apiPortTest {
		t.Errorf("expected cluster config node port before soft start to be %d but got %d", apiPortTest, beforeCfg.Config.KubernetesConfig.NodePort)
	}

	softStartArgs := []string{"start", "-p", profile, "--alsologtostderr", "-v=8"}
	c := exec.CommandContext(ctx, Target(), softStartArgs...)
	rr, err := Run(t, c)
	if err != nil {
		t.Errorf("failed to soft start minikube. args %q: %v", rr.Command(), err)
	}
	t.Logf("soft start took %s for %q cluster.", time.Since(start), profile)

	afterCfg, err := config.LoadProfile(profile)
	if err != nil {
		t.Errorf("error reading cluster config after soft start: %v", err)
	}

	if afterCfg.Config.KubernetesConfig.NodePort != apiPortTest {
		t.Errorf("expected node port in the config not change after soft start. exepceted node port to be %d but got %d.", apiPortTest, afterCfg.Config.KubernetesConfig.NodePort)
	}

	t.Run("Audit", func(t *testing.T) {
		got, err := auditContains(profile)
		if err != nil {
			t.Fatalf("failed to check audit log: %v", err)
		}
		if !got {
			t.Errorf("audit.json does not contain the profile %q", profile)
		}
	})
}

// validateKubeContext asserts that kubectl is properly configured (race-condition prone!)
func validateKubeContext(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "config", "current-context"))
	if err != nil {
		t.Errorf("failed to get current-context. args %q : %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Stdout.String(), profile) {
		t.Errorf("expected current-context = %q, but got *%q*", profile, rr.Stdout.String())
	}
}

// validateKubectlGetPods asserts that `kubectl get pod -A` returns non-zero content
func validateKubectlGetPods(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "po", "-A"))
	if err != nil {
		t.Errorf("failed to get kubectl pods: args %q : %v", rr.Command(), err)
	}
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

	// Must set the profile so that it knows what version of Kubernetes to use
	kubectlArgs := []string{"-p", profile, "kubectl", "--", "--context", profile, "get", "pods"}
	rr, err := Run(t, exec.CommandContext(ctx, Target(), kubectlArgs...))
	if err != nil {
		t.Fatalf("failed to get pods. args %q: %v", rr.Command(), err)
	}
}

// validateMinikubeKubectlDirectCall validates that calling minikube's kubectl
func validateMinikubeKubectlDirectCall(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	dir := filepath.Dir(Target())
	dstfn := filepath.Join(dir, "kubectl")
	err := os.Link(Target(), dstfn)

	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dstfn) // clean up

	kubectlArgs := []string{"--context", profile, "get", "pods"}
	rr, err := Run(t, exec.CommandContext(ctx, dstfn, kubectlArgs...))
	if err != nil {
		t.Fatalf("failed to run kubectl directly. args %q: %v", rr.Command(), err)
	}

}

func validateExtraConfig(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	start := time.Now()
	// The tests before this already created a profile, starting minikube with different --extra-config cmdline option.
	startArgs := []string{"start", "-p", profile, "--extra-config=apiserver.enable-admission-plugins=NamespaceAutoProvision"}
	c := exec.CommandContext(ctx, Target(), startArgs...)
	rr, err := Run(t, c)
	if err != nil {
		t.Errorf("failed to restart minikube. args %q: %v", rr.Command(), err)
	}
	t.Logf("restart took %s for %q cluster.", time.Since(start), profile)

	afterCfg, err := config.LoadProfile(profile)
	if err != nil {
		t.Errorf("error reading cluster config after soft start: %v", err)
	}

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

	if imgIds, ok := ids[image]; ok {
		if id, ok := imgIds[runtime.GOARCH]; ok {
			return id
		}
		panic(fmt.Sprintf("unexpected architecture for image %q: %v", image, runtime.GOARCH))
	}
	panic("unexpected image name: " + image)
}

// validateComponentHealth asserts that all Kubernetes components are healthy
func validateComponentHealth(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	// The ComponentStatus API is deprecated in v1.19, so do the next closest thing.
	found := map[string]bool{
		"etcd":                    false,
		"kube-apiserver":          false,
		"kube-controller-manager": false,
		"kube-scheduler":          false,
	}

	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "get", "po", "-l", "tier=control-plane", "-n", "kube-system", "-o=json"))
	if err != nil {
		t.Fatalf("failed to get components. args %q: %v", rr.Command(), err)
	}
	cs := api.PodList{}
	d := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes()))
	if err := d.Decode(&cs); err != nil {
		t.Fatalf("failed to decode kubectl json output: args %q : %v", rr.Command(), err)
	}

	for _, i := range cs.Items {
		for _, l := range i.Labels {
			t.Logf("%s phase: %s", l, i.Status.Phase)
			_, ok := found[l]
			if ok {
				found[l] = true
				if i.Status.Phase != "Running" {
					t.Errorf("%s is not Running: %+v", l, i.Status)
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

func validateStatusCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status"))
	if err != nil {
		t.Errorf("failed to run minikube status. args %q : %v", rr.Command(), err)
	}

	// Custom format
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-f", "host:{{.Host}},kublet:{{.Kubelet}},apiserver:{{.APIServer}},kubeconfig:{{.Kubeconfig}}"))
	if err != nil {
		t.Errorf("failed to run minikube status with custom format: args %q: %v", rr.Command(), err)
	}
	re := `host:([A-z]+),kublet:([A-z]+),apiserver:([A-z]+),kubeconfig:([A-z]+)`
	match, _ := regexp.MatchString(re, rr.Stdout.String())
	if !match {
		t.Errorf("failed to match regex %q for minikube status with custom format. args %q. output: %s", re, rr.Command(), rr.Output())
	}

	// Json output
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "status", "-o", "json"))
	if err != nil {
		t.Errorf("failed to run minikube status with json output. args %q : %v", rr.Command(), err)
	}
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

	args := []string{"dashboard", "--url", "-p", profile, "--alsologtostderr", "-v=1"}
	ss, err := Start(t, exec.CommandContext(ctx, Target(), args...))
	if err != nil {
		t.Errorf("failed to run minikube dashboard. args %q : %v", args, err)
	}
	defer func() {
		ss.Stop(t)
	}()

	start := time.Now()
	s, err := ReadLineWithTimeout(ss.Stdout, Seconds(300))
	if err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("failed to read url within %s: %v\noutput: %q\n", time.Since(start), err, s)
		}
		t.Fatalf("failed to read url within %s: %v\noutput: %q\n", time.Since(start), err, s)
	}

	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil {
		t.Fatalf("failed to parse %q: %v", s, err)
	}

	resp, err := retryablehttp.Get(u.String())
	if err != nil {
		t.Fatalf("failed to http get %q: %v\nresponse: %+v", u.String(), err, resp)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("failed to read http response body from dashboard %q: %v", u.String(), err)
		}
		t.Errorf("%s returned status code %d, expected %d.\nbody:\n%s", u, resp.StatusCode, http.StatusOK, body)
	}
}

// validateDryRun asserts that the dry-run mode quickly exits with the right code
func validateDryRun(ctx context.Context, t *testing.T, profile string) {
	// dry-run mode should always be able to finish quickly (<5s)
	mctx, cancel := context.WithTimeout(ctx, Seconds(5))
	defer cancel()

	// Too little memory!
	startArgs := append([]string{"start", "-p", profile, "--dry-run", "--memory", "250MB", "--alsologtostderr"}, StartArgs()...)
	c := exec.CommandContext(mctx, Target(), startArgs...)
	rr, err := Run(t, c)

	wantCode := reason.ExInsufficientMemory
	if rr.ExitCode != wantCode {
		t.Errorf("dry-run(250MB) exit code = %d, wanted = %d: %v", rr.ExitCode, wantCode, err)
	}

	dctx, cancel := context.WithTimeout(ctx, Seconds(5))
	defer cancel()
	startArgs = append([]string{"start", "-p", profile, "--dry-run", "--alsologtostderr", "-v=1"}, StartArgs()...)
	c = exec.CommandContext(dctx, Target(), startArgs...)
	rr, err = Run(t, c)
	if rr.ExitCode != 0 || err != nil {
		t.Errorf("dry-run exit code = %d, wanted = %d: %v", rr.ExitCode, 0, err)
	}
}

// validateCacheCmd tests functionality of cache command (cache add, delete, list)
func validateCacheCmd(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	if NoneDriver() {
		t.Skipf("skipping: cache unsupported by none")
	}

	t.Run("cache", func(t *testing.T) {
		t.Run("add_remote", func(t *testing.T) {
			for _, img := range []string{"k8s.gcr.io/pause:3.1", "k8s.gcr.io/pause:3.3", "k8s.gcr.io/pause:latest"} {
				rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "cache", "add", img))
				if err != nil {
					t.Errorf("failed to 'cache add' remote image %q. args %q err %v", img, rr.Command(), err)
				}
			}
		})

		t.Run("add_local", func(t *testing.T) {
			if GithubActionRunner() && runtime.GOOS == "darwin" {
				t.Skipf("skipping this test because Docker can not run in macos on github action free version. https://github.community/t/is-it-possible-to-install-and-configure-docker-on-macos-runner/16981")
			}

			_, err := exec.LookPath(oci.Docker)
			if err != nil {
				t.Skipf("docker is not installed, skipping local image test")
			}

			dname, err := ioutil.TempDir("", profile)
			if err != nil {
				t.Fatalf("Cannot create temp dir: %v", err)
			}

			message := []byte("FROM scratch\nADD Dockerfile /x")
			err = ioutil.WriteFile(filepath.Join(dname, "Dockerfile"), message, 0644)
			if err != nil {
				t.Fatalf("unable to write Dockerfile: %v", err)
			}

			img := "minikube-local-cache-test:" + profile
			_, err = Run(t, exec.CommandContext(ctx, "docker", "build", "-t", img, dname))
			if err != nil {
				t.Skipf("failed to build docker image, skipping local test: %v", err)
			}

			rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "cache", "add", img))
			if err != nil {
				t.Errorf("failed to 'cache add' local image %q. args %q err %v", img, rr.Command(), err)
			}
		})

		t.Run("delete_k8s.gcr.io/pause:3.3", func(t *testing.T) {
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "cache", "delete", "k8s.gcr.io/pause:3.3"))
			if err != nil {
				t.Errorf("failed to delete image k8s.gcr.io/pause:3.3 from cache. args %q: %v", rr.Command(), err)
			}
		})

		t.Run("list", func(t *testing.T) {
			rr, err := Run(t, exec.CommandContext(ctx, Target(), "cache", "list"))
			if err != nil {
				t.Errorf("failed to do cache list. args %q: %v", rr.Command(), err)
			}
			if !strings.Contains(rr.Output(), "k8s.gcr.io/pause") {
				t.Errorf("expected 'cache list' output to include 'k8s.gcr.io/pause' but got: ***%s***", rr.Output())
			}
			if strings.Contains(rr.Output(), "k8s.gcr.io/pause:3.3") {
				t.Errorf("expected 'cache list' output not to include k8s.gcr.io/pause:3.3 but got: ***%s***", rr.Output())
			}
		})

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

		t.Run("cache_reload", func(t *testing.T) { // deleting image inside minikube node manually and expecting reload to bring it back
			img := "k8s.gcr.io/pause:latest"
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
			for _, img := range []string{"k8s.gcr.io/pause:3.1", "k8s.gcr.io/pause:latest"} {
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

// validateLogsCmd asserts basic "logs" command functionality
func validateLogsCmd(ctx context.Context, t *testing.T, profile string) {
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "logs"))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}
	expectedWords := []string{"apiserver", "Linux", "kubelet"}
	switch ContainerRuntime() {
	case "docker":
		expectedWords = append(expectedWords, "Docker")
	case "containerd":
		expectedWords = append(expectedWords, "containerd")
	case "crio":
		expectedWords = append(expectedWords, "crio")
	}

	for _, word := range expectedWords {
		if !strings.Contains(rr.Stdout.String(), word) {
			t.Errorf("expected minikube logs to include word: -%q- but got \n***%s***\n", word, rr.Output())
		}
	}
}

// validateProfileCmd asserts "profile" command functionality
func validateProfileCmd(ctx context.Context, t *testing.T, profile string) {
	t.Run("profile_not_create", func(t *testing.T) {
		// Profile command should not create a nonexistent profile
		nonexistentProfile := "lis"
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", nonexistentProfile))
		if err != nil {
			t.Errorf("%s failed: %v", rr.Command(), err)
		}
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

	t.Run("profile_list", func(t *testing.T) {
		// List profiles
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list"))
		if err != nil {
			t.Errorf("failed to list profiles: args %q : %v", rr.Command(), err)
		}

		// Table output
		listLines := strings.Split(strings.TrimSpace(rr.Stdout.String()), "\n")
		profileExists := false
		for i := 3; i < (len(listLines) - 1); i++ {
			profileLine := listLines[i]
			if strings.Contains(profileLine, profile) {
				profileExists = true
				break
			}
		}
		if !profileExists {
			t.Errorf("expected 'profile list' output to include %q but got *%q*. args: %q", profile, rr.Stdout.String(), rr.Command())
		}
	})

	t.Run("profile_json_output", func(t *testing.T) {
		// Json output
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "--output", "json"))
		if err != nil {
			t.Errorf("failed to list profiles with json format. args %q: %v", rr.Command(), err)
		}
		var jsonObject map[string][]map[string]interface{}
		err = json.Unmarshal(rr.Stdout.Bytes(), &jsonObject)
		if err != nil {
			t.Errorf("failed to decode json from profile list: args %q: %v", rr.Command(), err)
		}
		validProfiles := jsonObject["valid"]
		profileExists := false
		for _, profileObject := range validProfiles {
			if profileObject["Name"] == profile {
				profileExists = true
				break
			}
		}
		if !profileExists {
			t.Errorf("expected the json of 'profile list' to include %q but got *%q*. args: %q", profile, rr.Stdout.String(), rr.Command())
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

	var rr *RunResult
	var err error
	// k8s.gcr.io/echoserver is not multi-arch
	if arm64Platform() {
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "deployment", "hello-node", "--image=k8s.gcr.io/echoserver-arm:1.8"))
	} else {
		rr, err = Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "create", "deployment", "hello-node", "--image=k8s.gcr.io/echoserver:1.8"))
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

	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "service", "list"))
	if err != nil {
		t.Errorf("failed to do service list. args %q : %v", rr.Command(), err)
	}
	if !strings.Contains(rr.Stdout.String(), "hello-node") {
		t.Errorf("expected 'service list' to contain *hello-node* but got -%q-", rr.Stdout.String())
	}

	if NeedsPortForward() {
		t.Skipf("test is broken for port-forwarded drivers: https://github.com/kubernetes/minikube/issues/7383")
	}

	// Test --https --url mode
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "service", "--namespace=default", "--https", "--url", "hello-node"))
	if err != nil {
		t.Fatalf("failed to get service url. args %q : %v", rr.Command(), err)
	}
	if rr.Stderr.String() != "" {
		t.Errorf("expected stderr to be empty but got *%q* . args %q", rr.Stderr, rr.Command())
	}

	endpoint := strings.TrimSpace(rr.Stdout.String())
	t.Logf("found endpoint: %s", endpoint)

	u, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("failed to parse service url endpoint %q: %v", endpoint, err)
	}
	if u.Scheme != "https" {
		t.Errorf("expected scheme for %s to be 'https' but got %q", endpoint, u.Scheme)
	}

	// Test --format=IP
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "service", "hello-node", "--url", "--format={{.IP}}"))
	if err != nil {
		t.Errorf("failed to get service url with custom format. args %q: %v", rr.Command(), err)
	}
	if strings.TrimSpace(rr.Stdout.String()) != u.Hostname() {
		t.Errorf("expected 'service --format={{.IP}}' output to be -%q- but got *%q* . args %q.", u.Hostname(), rr.Stdout.String(), rr.Command())
	}

	// Test a regular URL
	rr, err = Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "service", "hello-node", "--url"))
	if err != nil {
		t.Errorf("failed to get service url. args: %q: %v", rr.Command(), err)
	}

	endpoint = strings.TrimSpace(rr.Stdout.String())
	t.Logf("found endpoint for hello-node: %s", endpoint)

	u, err = url.Parse(endpoint)
	if err != nil {
		t.Fatalf("failed to parse %q: %v", endpoint, err)
	}

	if u.Scheme != "http" {
		t.Fatalf("expected scheme to be -%q- got scheme: *%q*", "http", u.Scheme)
	}

	t.Logf("Attempting to fetch %s ...", endpoint)

	fetch := func() error {
		resp, err := http.Get(endpoint)
		if err != nil {
			t.Logf("error fetching %s: %v", endpoint, err)
			return err
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
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

	// Table output
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "addons", "list"))
	if err != nil {
		t.Errorf("failed to do addon list: args %q : %v", rr.Command(), err)
	}
	for _, a := range []string{"dashboard", "ingress", "ingress-dns"} {
		if !strings.Contains(rr.Output(), a) {
			t.Errorf("expected 'addon list' output to include -%q- but got *%s*", a, rr.Output())
		}
	}

	// Json output
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

	// testing hostname as well because testing something like "minikube ssh echo" could be confusing
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

// validateMySQL validates a minimalist MySQL deployment
func validateMySQL(ctx context.Context, t *testing.T, profile string) {
	if arm64Platform() {
		t.Skip("arm64 is not supported by mysql. Skip the test. See https://github.com/kubernetes/minikube/issues/10144")
	}

	defer PostMortemLogs(t, profile)

	rr, err := Run(t, exec.CommandContext(ctx, "kubectl", "--context", profile, "replace", "--force", "-f", filepath.Join(*testdataDir, "mysql.yaml")))
	if err != nil {
		t.Fatalf("failed to kubectl replace mysql: args %q failed: %v", rr.Command(), err)
	}

	names, err := PodWait(ctx, t, profile, "default", "app=mysql", Minutes(10))
	if err != nil {
		t.Fatalf("failed waiting for mysql pod: %v", err)
	}

	// Retry, as mysqld first comes up without users configured. Scan for names in case of a reschedule.
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

// localTestCertPath is where the test file will be synced into the VM
func localTestCertPath() string {
	return filepath.Join(localpath.MiniPath(), "/certs", testCert())
}

// localEmptyCertPath is where the test file will be synced into the VM
func localEmptyCertPath() string {
	return filepath.Join(localpath.MiniPath(), "/certs", fmt.Sprintf("%d_empty.pem", os.Getpid()))
}

// Copy extra file into minikube home folder for file sync test
func setupFileSync(ctx context.Context, t *testing.T, profile string) {
	p := localSyncTestPath()
	t.Logf("local sync path: %s", p)
	syncFile := filepath.Join(*testdataDir, "sync.test")
	err := copy.Copy(syncFile, p)
	if err != nil {
		t.Fatalf("failed to copy testdata/sync.test: %v", err)
	}

	testPem := filepath.Join(*testdataDir, "minikube_test.pem")

	// Write to a temp file for an atomic write
	tmpPem := localTestCertPath() + ".pem"
	if err := copy.Copy(testPem, tmpPem); err != nil {
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

	// Create an empty file just to mess with people
	if _, err := os.Create(localEmptyCertPath()); err != nil {
		t.Fatalf("create failed: %v", err)
	}
}

// validateFileSync to check existence of the test file
func validateFileSync(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	if NoneDriver() {
		t.Skipf("skipping: ssh unsupported by none")
	}

	vp := vmSyncTestPath()
	t.Logf("Checking for existence of %s within VM", vp)
	rr, err := Run(t, exec.CommandContext(ctx, Target(), "-p", profile, "ssh", fmt.Sprintf("sudo cat %s", vp)))
	if err != nil {
		t.Errorf("%s failed: %v", rr.Command(), err)
	}
	got := rr.Stdout.String()
	t.Logf("file sync test content: %s", got)

	syncFile := filepath.Join(*testdataDir, "sync.test")
	expected, err := ioutil.ReadFile(syncFile)
	if err != nil {
		t.Errorf("failed to read test file 'testdata/sync.test' : %v", err)
	}

	if diff := cmp.Diff(string(expected), got); diff != "" {
		t.Errorf("/etc/sync.test content mismatch (-want +got):\n%s", diff)
	}
}

// validateCertSync to check existence of the test certificate
func validateCertSync(ctx context.Context, t *testing.T, profile string) {
	defer PostMortemLogs(t, profile)

	if NoneDriver() {
		t.Skipf("skipping: ssh unsupported by none")
	}

	testPem := filepath.Join(*testdataDir, "minikube_test.pem")
	want, err := ioutil.ReadFile(testPem)
	if err != nil {
		t.Errorf("test file not found: %v", err)
	}

	// Check both the installed & reference certs (they should be symlinked)
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
		got := strings.Replace(rr.Stdout.String(), "\r", "", -1)
		if diff := cmp.Diff(string(want), got); diff != "" {
			t.Errorf("failed verify pem file. minikube_test.pem -> %s mismatch (-want +got):\n%s", vp, diff)
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
				tf, err := ioutil.TempFile("", "kubeconfig")
				if err != nil {
					t.Fatal(err)
				}

				if err := ioutil.WriteFile(tf.Name(), tc.kubeconfig, 0644); err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() {
					os.Remove(tf.Name())
				})

				c.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", tf.Name()))
			}

			rr, err := Run(t, c)
			if err != nil {
				t.Errorf("failed to run minikube update-context: args %q: %v", rr.Command(), err)
			}

			if !bytes.Contains(rr.Stdout.Bytes(), tc.want) {
				t.Errorf("update-context: got=%q, want=*%q*", rr.Stdout.Bytes(), tc.want)
			}
		})
	}
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
