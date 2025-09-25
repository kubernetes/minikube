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

	"github.com/blang/semver/v4"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/registry"
)

var docURL = "https://minikube.sigs.k8s.io/docs/drivers/podman/"

// minReqPodmanVer is required the minimum version of podman to be installed for podman driver.
var minReqPodmanVer = semver.Version{Major: 4, Minor: 9, Patch: 0}

func init() {
	priority := registry.Default
	if runtime.GOOS != "linux" {
		// requires external VM set up
		priority = registry.Experimental
	}
	// Staged rollout for default:
	// - Linux (sudo podman)
	// - macOS (podman-remote)
	// - Windows (podman-remote)

	if err := registry.Register(registry.DriverDef{
		Name:     driver.Podman,
		Config:   configure,
		Init:     func() drivers.Driver { return kic.NewDriver(kic.Config{OCIBinary: oci.Podman}) },
		Status:   status,
		Default:  true,
		Priority: priority,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	mounts := make([]oci.Mount, len(cc.ContainerVolumeMounts))
	for i, spec := range cc.ContainerVolumeMounts {
		var err error
		mounts[i], err = oci.ParseMountString(spec)
		if err != nil {
			return nil, err
		}
	}

	extraArgs := []string{}

	for _, port := range cc.ExposedPorts {
		extraArgs = append(extraArgs, "-p", port)
	}

	return kic.NewDriver(kic.Config{
		ClusterName:       cc.Name,
		MachineName:       config.MachineName(cc, n),
		StorePath:         localpath.MiniPath(),
		ImageDigest:       strings.Split(cc.KicBaseImage, "@")[0], // for podman does not support docker images references with both a tag and digest.
		Mounts:            mounts,
		CPU:               cc.CPUs,
		Memory:            cc.Memory,
		OCIBinary:         oci.Podman,
		APIServerPort:     cc.Nodes[0].Port,
		KubernetesVersion: cc.KubernetesConfig.KubernetesVersion,
		ContainerRuntime:  cc.KubernetesConfig.ContainerRuntime,
		ExtraArgs:         extraArgs,
		ListenAddress:     cc.ListenAddress,
		Subnet:            cc.Subnet,
	}), nil
}

func status() registry.State {
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
		cmd = exec.CommandContext(ctx, oci.Podman, "version", "--format", "{{.Version}}")
		cmd = oci.PrefixCmd(cmd, oci.WithSudoFlags("-k"))
		cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C") // sudo is localized
	}
	o, err := cmd.Output()
	output := strings.TrimSpace(string(o))
	if err == nil {
		klog.Infof("podman version: %s", output)

		v, err := semver.Make(output)
		if err != nil {
			return registry.State{Error: err, Installed: true, Running: true, Healthy: false, Fix: "Can't verify minimum required version for podman . See podman website for installation guide.", Doc: "https://podman.io/getting-started/installation.html"}
		}

		if v.LT(minReqPodmanVer) {
			out.WarningT(`The minimum required version for podman is "{{.minVersion}}". your version is "{{.currentVersion}}". minikube might not work. use at your own risk. To install latest version please see https://podman.io/getting-started/installation.html`,
				out.V{"minVersion": minReqPodmanVer.String(), "currentVersion": v.String()})
		}

		return registry.State{Installed: true, Healthy: true}
	}

	klog.Warningf("podman returned error: %v", err)

	// Basic timeout
	if ctx.Err() == context.DeadlineExceeded {
		return registry.State{Error: err, Installed: true, Running: false, Healthy: false, Fix: "Restart the Podman service", Doc: docURL}
	}

	username := "$USER"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		newErr := fmt.Errorf(`%q %v: %s`, strings.Join(cmd.Args, " "), exitErr, stderr)

		if strings.Contains(stderr, "a password is required") && runtime.GOOS == "linux" {
			return registry.State{Error: newErr, Installed: true, Healthy: false, Fix: fmt.Sprintf("Add your user to the 'sudoers' file: '%s ALL=(ALL) NOPASSWD: %s' , or run 'minikube config set rootless true'", username, podman), Doc: "https://podman.io"}
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
