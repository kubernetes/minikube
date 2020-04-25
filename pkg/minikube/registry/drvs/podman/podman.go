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
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

// minReqPodmanVer is required the mininum version of podman to be installed for podman driver.
var minReqPodmanVer = semver.Version{Major: 1, Minor: 7, Patch: 0}

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.Podman,
		Config:   configure,
		Init:     func() drivers.Driver { return kic.NewDriver(kic.Config{OCIPrefix: "sudo", OCIBinary: oci.Podman}) },
		Status:   status,
		Priority: registry.Experimental,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(cc config.ClusterConfig, n config.Node) (interface{}, error) {
	baseImage := viper.GetString("base-image")
	return kic.NewDriver(kic.Config{
		MachineName:   driver.MachineName(cc, n),
		StorePath:     localpath.MiniPath(),
		ImageDigest:   strings.Split(baseImage, "@")[0], // for podman does not support docker images references with both a tag and digest.
		CPU:           cc.CPUs,
		Memory:        cc.Memory,
		OCIPrefix:     "sudo",
		OCIBinary:     oci.Podman,
		APIServerPort: cc.Nodes[0].Port,
	}), nil
}

func status() registry.State {
	docURL := "https://minikube.sigs.k8s.io/docs/drivers/podman/"
	podman, err := exec.LookPath(oci.Podman)
	if err != nil {
		return registry.State{Error: err, Installed: false, Healthy: false, Fix: "Install Podman", Doc: docURL}
	}

	// Allow no more than 2 seconds for version command
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, oci.Podman, "version", "-f", "{{.Version}}")
	o, err := cmd.CombinedOutput()
	output := string(o)
	if err != nil {
		return registry.State{Error: err, Installed: true, Healthy: false, Fix: "Cant verify mininim required version for podman . See podman website for installation guide.", Doc: "https://podman.io/getting-started/installation.html"}
	}

	v, err := semver.Make(output)
	if err != nil {
		return registry.State{Error: err, Installed: true, Healthy: false, Fix: "Cant verify mininim required version for podman . See podman website for installation guide.", Doc: "https://podman.io/getting-started/installation.html"}
	}

	if v.LT(minReqPodmanVer) {
		glog.Warningf("Warning ! mininim required version for podman is %s. your version is %q. minikube might not work. use at your own risk. To install latest version please see https://podman.io/getting-started/installation.html ", minReqPodmanVer.String(), v.String())
	}

	// Allow no more than 3 seconds for querying state
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Run with sudo on linux (local), otherwise podman-remote (as podman)
	if runtime.GOOS == "linux" {
		cmd = exec.CommandContext(ctx, "sudo", "-n", oci.Podman, "info")
		cmd.Env = append(os.Environ(), "LANG=C", "LC_ALL=C") // sudo is localized
	} else {
		cmd = exec.CommandContext(ctx, oci.Podman, "info")
	}
	_, err = cmd.Output()
	if err == nil {
		return registry.State{Installed: true, Healthy: true}
	}

	glog.Warningf("podman returned error: %v", err)

	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		newErr := fmt.Errorf(`%q %v: %s`, strings.Join(cmd.Args, " "), exitErr, stderr)

		username := "$USER"
		if u, err := user.Current(); err == nil {
			username = u.Username
		}

		if strings.Contains(stderr, "a password is required") && runtime.GOOS == "linux" {
			return registry.State{Error: newErr, Installed: true, Healthy: false, Fix: fmt.Sprintf("Add your user to the 'sudoers' file: '%s ALL=(ALL) NOPASSWD: %s'", username, podman), Doc: "https://podman.io"}
		}

		// We don't have good advice, but at least we can provide a good error message
		return registry.State{Error: newErr, Installed: true, Healthy: false, Doc: docURL}
	}

	return registry.State{Error: err, Installed: true, Healthy: false, Doc: docURL}
}
