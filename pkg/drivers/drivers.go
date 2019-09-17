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
	"path"
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

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util"
)

const (
	driverKVMDownloadURL      = "https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-kvm2"
	driverHyperKitDownloadURL = "https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit"
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
		if err := fixPermissions(machPath); err != nil {
			return errors.Wrapf(err, "fixing permissions on %s", machPath)
		}
	}
	return nil
}

func fixPermissions(path string) error {
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
func InstallOrUpdate(driver, destination string, minikubeVersion semver.Version) error {
	_, err := exec.LookPath(driver)
	// if file driver doesn't exist, download it
	if err != nil {
		return download(driver, destination)
	}

	cmd := exec.Command(driver, "version")
	output, err := cmd.Output()
	// if driver doesnt support 'version', it is old, download it
	if err != nil {
		return download(driver, destination)
	}

	v := ExtractVMDriverVersion(string(output))

	// if the driver doesn't return any version, download it
	if len(v) == 0 {
		return download(driver, destination)
	}

	vmDriverVersion, err := semver.Make(v)
	if err != nil {
		return errors.Wrap(err, "can't parse driver version")
	}

	// if the current driver version is older, download newer
	if vmDriverVersion.LT(minikubeVersion) {
		return download(driver, destination)
	}

	return nil
}

func download(driver, destination string) error {
	// supports kvm2 and hyperkit
	if driver != "docker-machine-driver-kvm2" && driver != "docker-machine-driver-hyperkit" {
		return nil
	}

	out.T(out.FileDownload, "Downloading driver {{.driver}}:", out.V{"driver": driver})

	targetFilepath := path.Join(destination, driver)
	os.Remove(targetFilepath)

	var url string
	switch driver {
	case "docker-machine-driver-kvm2":
		url = driverKVMDownloadURL
	case "docker-machine-driver-hyperkit":
		url = driverHyperKitDownloadURL
	}

	opts := []getter.ClientOption{getter.WithProgress(util.DefaultProgressBar)}
	client := &getter.Client{
		Src:     url,
		Dst:     targetFilepath,
		Mode:    getter.ClientModeFile,
		Options: opts,
	}

	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "can't download driver %s from: %s", driver, url)
	}

	err := os.Chmod(targetFilepath, 0777)
	if err != nil {
		return errors.Wrap(err, "chmod error")
	}

	if driver == "docker-machine-driver-hyperkit" {
		err := setHyperKitPermissions(targetFilepath)
		if err != nil {
			return errors.Wrap(err, "setting hyperkit permission")
		}
	}

	return nil
}

// ExtractVMDriverVersion extracts the driver version.
// KVM and Hyperkit drivers support the 'version' command, that display the information as:
// version: vX.X.X
// commit: XXXX
// This method returns the version 'vX.X.X' or empty if the version isn't found.
func ExtractVMDriverVersion(s string) string {
	versionRegex := regexp.MustCompile(`version:(.*)`)
	matches := versionRegex.FindStringSubmatch(s)

	if len(matches) != 2 {
		return ""
	}

	v := strings.TrimSpace(matches[1])
	return strings.TrimPrefix(v, version.VersionPrefix)
}

func setHyperKitPermissions(driverPath string) error {
	msg := fmt.Sprintf("A new hyperkit driver was installed. It needs elevated permissions to run. The following commands will be executed:\n  sudo chown root:wheel %s\n  sudo chmod u+s %s", driverPath, driverPath)
	out.T(out.Permissions, msg, out.V{})

	cmd := exec.Command("sudo", "chown", "root:wheel", driverPath)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "chown root:wheel")
	}

	cmd = exec.Command("sudo", "chmod", "u+s", driverPath)
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(err, "chmod u+s")
	}

	return nil
}
