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

package auxdriver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver/v4"

	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/lock"
)

func newAuxUnthealthyError(path string) error {
	return fmt.Errorf(`failed to execute auxiliary version command "%s --version"`, path)
}

func newAuxNotFoundError(name, path string) error {
	return fmt.Errorf("auxiliary driver %s not found in path %s", name, path)
}

// ErrAuxDriverVersionCommandFailed indicates the aux driver 'version' command failed to run
var ErrAuxDriverVersionCommandFailed error

// ErrAuxDriverVersionNotinPath was not found in PATH
var ErrAuxDriverVersionNotinPath error

// InstallOrUpdate downloads driver if it is not present, or updates it if there's a newer version
// this is currently only used for hyperkit driver on macOS
func InstallOrUpdate(name string, directory string, interactive bool, autoUpdate bool) error {
	if name != driver.HyperKit {
		return nil
	}
	// v1.37.0 was the last release that we built hyperkit driver https://github.com/kubernetes/minikube/issues/21940
	// this will be used till we completely remove hyperkit driver support
	v, err := semver.Make("1.37.0")
	out.WarningT("Hyperkit driver will be removed in the next minikube release, we have other drivers that work on macOS such as docker or qemu, vfkit. Please consider switching to one of them. For more information, please visit: https://minikube.sigs.k8s.io/docs/drivers/hyperkit/")
	if err != nil {
		return fmt.Errorf("can't parse version: %w", err)
	}

	executable := fmt.Sprintf("docker-machine-driver-%s", name)

	// Lock before we check for existence to avoid thundering herd issues
	spec := lock.PathMutexSpec(executable)
	spec.Timeout = 10 * time.Minute
	klog.Infof("acquiring lock: %+v", spec)
	releaser, err := lock.Acquire(spec)
	if err != nil {
		return fmt.Errorf("unable to acquire lock for %+v: %w", spec, err)
	}
	defer releaser.Release()

	exists := driverExists(executable)
	path, err := validateDriver(executable, minAcceptableDriverVersion(name, v))
	if !exists || (err != nil && autoUpdate) {
		klog.Warningf("%s: %v", executable, err)
		path = filepath.Join(directory, executable)
		if err := download.Driver(executable, path, v); err != nil {
			return err
		}
	}
	if err := fixDriverPermissions(name, path, interactive); err != nil {
		return err
	}

	if _, err := validateDriver(executable, minAcceptableDriverVersion(name, v)); err != nil {
		return err
	}
	return nil
}

// fixDriverPermissions fixes the permissions on a driver
func fixDriverPermissions(name string, path string, interactive bool) error {
	if name != driver.HyperKit {
		return nil
	}

	// Using the find command for hyperkit is far easier than cross-platform uid checks in Go.
	stdout, err := exec.Command("find", path, "-uid", "0", "-perm", "4755").Output()
	klog.Infof("stdout: %s", stdout)
	if err == nil && strings.TrimSpace(string(stdout)) == path {
		klog.Infof("%s looks good", path)
		return nil
	}

	cmds := []*exec.Cmd{
		exec.Command("sudo", "chown", "root:wheel", path),
		exec.Command("sudo", "chmod", "u+s", path),
	}

	var example strings.Builder
	for _, c := range cmds {
		example.WriteString(fmt.Sprintf("    $ %s \n", strings.Join(c.Args, " ")))
	}

	out.Styled(style.Permissions, "The '{{.driver}}' driver requires elevated permissions. The following commands will be executed:\n\n{{ .example }}\n", out.V{"driver": name, "example": example.String()})
	for _, c := range cmds {
		testArgs := append([]string{"-n"}, c.Args[1:]...)
		test := exec.Command("sudo", testArgs...)
		klog.Infof("testing: %v", test.Args)
		if err := test.Run(); err != nil {
			klog.Infof("%v may require a password: %v", c.Args, err)
			if !interactive {
				return fmt.Errorf("%v requires a password, and --interactive=false", c.Args)
			}
		}
		klog.Infof("running: %v", c.Args)
		err := c.Run()
		if err != nil {
			return fmt.Errorf("%v: %w", c.Args, err)
		}
	}
	return nil
}

// validateDriver validates if a driver appears to be up-to-date and installed properly
func validateDriver(executable string, v semver.Version) (string, error) {
	klog.Infof("Validating %s, PATH=%s", executable, os.Getenv("PATH"))
	path, err := exec.LookPath(executable)
	if err != nil {
		klog.Warningf("driver not in path : %s, %v", path, err.Error())
		ErrAuxDriverVersionNotinPath = newAuxNotFoundError(executable, path)
		return path, ErrAuxDriverVersionNotinPath
	}

	cmd := exec.Command(path, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Warningf("%s failed: %v: %s", cmd, err, output)
		ErrAuxDriverVersionCommandFailed = newAuxUnthealthyError(path)
		return path, ErrAuxDriverVersionCommandFailed
	}

	ev := extractDriverVersion(string(output))
	if len(ev) == 0 {
		return path, fmt.Errorf("%s: unable to extract version from %q", executable, output)
	}

	driverVersion, err := semver.Make(ev)
	if err != nil {
		return path, fmt.Errorf("can't parse driver version: %w", err)
	}
	klog.Infof("%s version is %s", path, driverVersion)

	if driverVersion.LT(v) {
		return path, fmt.Errorf("%s is version %s, want %s", executable, driverVersion, v)
	}
	return path, nil
}

// extractDriverVersion extracts the driver version.
// KVM and Hyperkit drivers support the 'version' command, that display the information as:
// version: vX.X.X
// commit: XXXX
// This method returns the version 'vX.X.X' or empty if the version isn't found.
func extractDriverVersion(s string) string {
	versionRegex := regexp.MustCompile(`version:(.*)`)
	matches := versionRegex.FindStringSubmatch(s)

	if len(matches) != 2 {
		return ""
	}

	v := strings.TrimSpace(matches[1])
	return strings.TrimPrefix(v, "v")
}

func driverExists(driverName string) bool {
	_, err := exec.LookPath(driverName)
	return err == nil
}
