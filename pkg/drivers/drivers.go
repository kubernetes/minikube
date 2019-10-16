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

package drivers

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/blang/semver"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/version"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util"
)

// GetDiskPath returns the path of the machine disk image
func GetDiskPath(d *drivers.BaseDriver) string {
	return filepath.Join(d.ResolveStorePath("."), d.GetMachineName()+".rawdisk")
}

// CommonDriver is the common driver base class
type CommonDriver struct{}

// GetCreateFlags is not implemented yet
func (d *CommonDriver) GetCreateFlags() []mcnflag.Flag {
	return nil
}

// SetConfigFromFlags is not implemented yet
func (d *CommonDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}

func createRawDiskImage(sshKeyPath, diskPath string, diskSizeMb int) error {
	tarBuf, err := mcnutils.MakeDiskImage(sshKeyPath)
	if err != nil {
		return errors.Wrap(err, "make disk image")
	}

	file, err := os.OpenFile(diskPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "open")
	}
	defer file.Close()
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errors.Wrap(err, "seek")
	}

	if _, err := file.Write(tarBuf.Bytes()); err != nil {
		return errors.Wrap(err, "write tar")
	}
	if err := file.Close(); err != nil {
		return errors.Wrapf(err, "closing file %s", diskPath)
	}

	if err := os.Truncate(diskPath, int64(diskSizeMb*1000000)); err != nil {
		return errors.Wrap(err, "truncate")
	}
	return nil
}

func publicSSHKeyPath(d *drivers.BaseDriver) string {
	return d.GetSSHKeyPath() + ".pub"
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func Restart(d drivers.Driver) error {
	if err := d.Stop(); err != nil {
		return err
	}

	return d.Start()
}

// MakeDiskImage makes a boot2docker VM disk image.
func MakeDiskImage(d *drivers.BaseDriver, boot2dockerURL string, diskSize int) error {
	glog.Infof("Making disk image using store path: %s", d.StorePath)
	b2 := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2.CopyIsoToMachineDir(boot2dockerURL, d.MachineName); err != nil {
		return errors.Wrap(err, "copy iso to machine dir")
	}

	keyPath := d.GetSSHKeyPath()
	glog.Infof("Creating ssh key: %s...", keyPath)
	if err := ssh.GenerateSSHKey(keyPath); err != nil {
		return errors.Wrap(err, "generate ssh key")
	}

	diskPath := GetDiskPath(d)
	glog.Infof("Creating raw disk image: %s...", diskPath)
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		if err := createRawDiskImage(publicSSHKeyPath(d), diskPath, diskSize); err != nil {
			return errors.Wrapf(err, "createRawDiskImage(%s)", diskPath)
		}
		machPath := d.ResolveStorePath(".")
		if err := fixMachinePermissions(machPath); err != nil {
			return errors.Wrapf(err, "fixing permissions on %s", machPath)
		}
	}
	return nil
}

func fixMachinePermissions(path string) error {
	glog.Infof("Fixing permissions on %s ...", path)
	if err := os.Chown(path, syscall.Getuid(), syscall.Getegid()); err != nil {
		return errors.Wrap(err, "chown dir")
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return errors.Wrap(err, "read dir")
	}
	for _, f := range files {
		fp := filepath.Join(path, f.Name())
		if err := os.Chown(fp, syscall.Getuid(), syscall.Getegid()); err != nil {
			return errors.Wrap(err, "chown file")
		}
	}
	return nil
}

// InstallOrUpdate downloads driver if it is not present, or updates it if there's a newer version
func InstallOrUpdate(driver string, directory string, v semver.Version, interactive bool, autoUpdate bool) error {
	if driver != constants.DriverKvm2 && driver != constants.DriverHyperkit {
		return nil
	}

	executable := fmt.Sprintf("docker-machine-driver-%s", driver)
	exists := driverExists(executable)
	path, err := validateDriver(executable, v)
	if !exists || (err != nil && autoUpdate) {
		glog.Warningf("%s: %v", executable, err)
		path = filepath.Join(directory, executable)
		derr := download(executable, path, v)
		if derr != nil {
			return derr
		}
	}
	return fixDriverPermissions(driver, path, interactive)
}

// fixDriverPermissions fixes the permissions on a driver
func fixDriverPermissions(driver string, path string, interactive bool) error {
	// This method only supports hyperkit so far (because it's complicated)
	if driver != constants.DriverHyperkit {
		return nil
	}

	// Using the find command for hyperkit is far easier than cross-platform uid checks in Go.
	stdout, err := exec.Command("find", path, "-uid", "0", "-perm", "4755").Output()
	glog.Infof("stdout: %s", stdout)
	if err == nil && strings.TrimSpace(string(stdout)) == path {
		glog.Infof("%s looks good", path)
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

	out.T(out.Permissions, "The '{{.driver}}' driver requires elevated permissions. The following commands will be executed:\n\n{{ .example }}\n", out.V{"driver": driver, "example": example.String()})
	for _, c := range cmds {
		testArgs := append([]string{"-n"}, c.Args[1:]...)
		test := exec.Command("sudo", testArgs...)
		glog.Infof("testing: %v", test.Args)
		if err := test.Run(); err != nil {
			glog.Infof("%v may require a password: %v", c.Args, err)
			if !interactive {
				return fmt.Errorf("%v requires a password, and --interactive=false", c.Args)
			}
		}
		glog.Infof("running: %v", c.Args)
		err := c.Run()
		if err != nil {
			return errors.Wrapf(err, "%v", c.Args)
		}
	}
	return nil
}

// validateDriver validates if a driver appears to be up-to-date and installed properly
func validateDriver(driver string, v semver.Version) (string, error) {
	glog.Infof("Validating %s, PATH=%s", driver, os.Getenv("PATH"))
	path, err := exec.LookPath(driver)
	if err != nil {
		return path, err
	}

	output, err := exec.Command(path, "version").Output()
	if err != nil {
		return path, err
	}

	ev := extractVMDriverVersion(string(output))
	if len(ev) == 0 {
		return path, fmt.Errorf("%s: unable to extract version from %q", driver, output)
	}

	vmDriverVersion, err := semver.Make(ev)
	if err != nil {
		return path, errors.Wrap(err, "can't parse driver version")
	}
	if vmDriverVersion.LT(v) {
		return path, fmt.Errorf("%s is version %s, want %s", driver, vmDriverVersion, v)
	}
	return path, nil
}

func driverExists(driver string) bool {
	_, err := exec.LookPath(driver)
	return err == nil
}

func driverWithChecksumURL(driver string, v semver.Version) string {
	base := fmt.Sprintf("https://github.com/kubernetes/minikube/releases/download/v%s/%s", v, driver)
	return fmt.Sprintf("%s?checksum=file:%s.sha256", base, base)
}

// download an arbitrary driver
func download(driver string, destination string, v semver.Version) error {
	out.T(out.FileDownload, "Downloading driver {{.driver}}:", out.V{"driver": driver})
	os.Remove(destination)
	url := driverWithChecksumURL(driver, v)
	client := &getter.Client{
		Src:     url,
		Dst:     destination,
		Mode:    getter.ClientModeFile,
		Options: []getter.ClientOption{getter.WithProgress(util.DefaultProgressBar)},
	}

	glog.Infof("Downloading: %+v", client)
	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}
	// Give downloaded drivers a baseline decent file permission
	return os.Chmod(destination, 0755)
}

// extractVMDriverVersion extracts the driver version.
// KVM and Hyperkit drivers support the 'version' command, that display the information as:
// version: vX.X.X
// commit: XXXX
// This method returns the version 'vX.X.X' or empty if the version isn't found.
func extractVMDriverVersion(s string) string {
	versionRegex := regexp.MustCompile(`version:(.*)`)
	matches := versionRegex.FindStringSubmatch(s)

	if len(matches) != 2 {
		return ""
	}

	v := strings.TrimSpace(matches[1])
	return strings.TrimPrefix(v, version.VersionPrefix)
}
