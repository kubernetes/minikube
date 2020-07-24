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

package podman

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/golang/glog"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/registry"
)

// minReqPodmanVer is required the minimum version of podman to be installed for podman driver.
var minReqPodmanVer = semver.Version{Major: 1, Minor: 7, Patch: 0}

// podmanVerTwo is required to exit with an error when podman v2 driver is currently installed because it is not supported yet.
var podmanVerTwo = semver.Version{Major: 2, Minor: 0, Patch: 0}

func init() {
	priority := registry.Experimental
	// Staged rollout for default:
	// - Linux
	// - macOS (podman-remote)
	// - Windows (podman-remote)

	if err := registry.Register(registry.DriverDef{
		Name:     driver.Podman,
		Config:   configure,
		Init:     func() drivers.Driver { return kic.NewDriver(kic.Config{OCIBinary: oci.Podman}) },
		Status:   status,
		Priority: priority,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	return kic.NewDriver(kic.Config{
		MachineName:       driver.MachineName(cc, n),
		StorePath:         localpath.MiniPath(),
		ImageDigest:       strings.Split(cc.KicBaseImage, "@")[0], // for podman does not support docker images references with both a tag and digest.
		CPU:               cc.CPUs,
		Memory:            cc.Memory,
		OCIBinary:         oci.Podman,
		APIServerPort:     cc.Nodes[0].Port,
		KubernetesVersion: cc.KubernetesConfig.KubernetesVersion,
		ContainerRuntime:  cc.KubernetesConfig.ContainerRuntime,
	}), nil
}

func status() registry.State {
	docURL := "https://minikube.sigs.k8s.io/docs/drivers/podman/"
	if runtime.GOARCH != "amd64" {
		return registry.State{Error: fmt.Errorf("podman driver is not supported on %q systems yet", runtime.GOARCH), Installed: false, Healthy: false, Fix: "Try other drivers", Doc: docURL}
	}

	podman, err := exec.LookPath(oci.Podman)
	if err != nil {
		return registry.State{Error: err, Installed: false, Healthy: false, Fix: "Install Podman", Doc: docURL}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	// Quickly returns an error code if service is not running
	cmd := exec.CommandContext(ctx, oci.Podman, "version", "--format", "{{.Server.Version}}")
	// Run with sudo on linux (local), otherwise podman-remote (as podman)
	if runtime.GOOS == "linux" {
		cmd = exec.CommandContext(ctx, "sudo", "-k", "-n", oci.Podman, "version", "--format", "{{.Version}}")
		cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C") // sudo is localized
	}
	o, err := cmd.Output()
	output := strings.TrimSpace(string(o))
	if err == nil {
		glog.Infof("podman version: %s", output)

		v, err := semver.Make(output)
		if err != nil {
			return registry.State{Error: err, Installed: true, Healthy: false, Fix: "Cant verify minimum required version for podman . See podman website for installation guide.", Doc: "https://podman.io/getting-started/installation.html"}
		}

		if v.LT(minReqPodmanVer) {
			out.WarningT(`The minimum required version for podman is "{{.minVersion}}". your version is "{{.currentVersion}}". minikube might not work. use at your own risk. To install latest version please see https://podman.io/getting-started/installation.html`,
				out.V{"minVersion": minReqPodmanVer.String(), "currentVersion": v.String()})
		} else if v.GTE(podmanVerTwo) {
			out.WarningT(`Using podman 2 is not supported yet. your version is "{{.currentVersion}}". minikube might not work. use at your own risk.`,
				out.V{"currentVersion": v.String()})
		}

		return registry.State{Installed: true, Healthy: true}
	}

	glog.Warningf("podman returned error: %v", err)

	// Basic timeout
	if ctx.Err() == context.DeadlineExceeded {
		return registry.State{Error: err, Installed: true, Healthy: false, Fix: "Restart the Podman service", Doc: docURL}
	}

	username := "$USER"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		newErr := fmt.Errorf(`%q %v: %s`, strings.Join(cmd.Args, " "), exitErr, stderr)

		if strings.Contains(stderr, "a password is required") && runtime.GOOS == "linux" {
			return registry.State{Error: newErr, Installed: true, Healthy: false, Fix: fmt.Sprintf("Add your user to the 'sudoers' file: '%s ALL=(ALL) NOPASSWD: %s'", username, podman), Doc: "https://podman.io"}
		}

		// Typical low-level errors from running podman-remote:
		// - local: "dial unix /run/podman/io.podman: connect: no such file or directory"
		// - remote: "unexpected EOF" (ssh varlink isn't so great at handling rejections)

		if strings.Contains(stderr, "could not get runtime") || strings.Contains(stderr, "Unable to obtain server version information") {
			return registry.State{Error: newErr, Installed: true, Healthy: false, Fix: "Start the Podman service", Doc: docURL}
		}

		// We don't have good advice, but at least we can provide a good error message
		return registry.State{Error: newErr, Installed: true, Healthy: false, Doc: docURL}
	}

	return registry.State{Error: err, Installed: true, Healthy: false, Doc: docURL}
}
