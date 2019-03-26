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
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/golang/glog"
	"github.com/pkg/errors"
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
		return err
	}

	file, err := os.OpenFile(diskPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Seek(0, os.SEEK_SET)

	if _, err := file.Write(tarBuf.Bytes()); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return errors.Wrapf(err, "closing file %s", diskPath)
	}

	if err := os.Truncate(diskPath, int64(diskSizeMb*1000000)); err != nil {
		return err
	}
	return nil
}

func publicSSHKeyPath(d *drivers.BaseDriver) string {
	return d.GetSSHKeyPath() + ".pub"
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func Restart(d drivers.Driver) error {
	for _, f := range []func() error{d.Stop, d.Start} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

// MakeDiskImage makes a boot2docker VM disk image.
func MakeDiskImage(d *drivers.BaseDriver, boot2dockerURL string, diskSize int) error {
	//TODO(r2d4): rewrite this, not using b2dutils
	b2dutils := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2dutils.CopyIsoToMachineDir(boot2dockerURL, d.MachineName); err != nil {
		return errors.Wrap(err, "Error copying ISO to machine dir")
	}

	glog.Info("Creating ssh key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	glog.Info("Creating raw disk image...")
	diskPath := GetDiskPath(d)
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		if err := createRawDiskImage(publicSSHKeyPath(d), diskPath, diskSize); err != nil {
			return err
		}
		if err := fixPermissions(d.ResolveStorePath(".")); err != nil {
			return err
		}
	}
	return nil
}

func fixPermissions(path string) error {
	os.Chown(path, syscall.Getuid(), syscall.Getegid())
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		fp := filepath.Join(path, f.Name())
		if err := os.Chown(fp, syscall.Getuid(), syscall.Getegid()); err != nil {
			return err
		}
	}
	return nil
}
