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

package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"k8s.io/minikube/pkg/libmachine/diagnostics"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/ssh"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/util"
)

// This file is for common code shared among internal machine drivers
// Code here should not be called from within minikube

// GetDiskPath returns the path of the machine disk image
func GetDiskPath(d *drivers.BaseDriver) string {
	return filepath.Join(d.ResolveStorePath("."), d.GetMachineName()+".rawdisk")
}

// ExtraDiskPath returns the path of an additional disk suffixed with an ID.
func ExtraDiskPath(d *drivers.BaseDriver, diskID int) string {
	file := fmt.Sprintf("%s-%d.rawdisk", d.GetMachineName(), diskID)
	return filepath.Join(d.ResolveStorePath("."), file)
}

// CreateRawDisk creates a new raw disk image.
//
// Example usage:
//
//	path := ExtraDiskPath(baseDriver, diskID)
//	err := CreateRawDisk(path, baseDriver.DiskSize)
func CreateRawDisk(diskPath string, sizeMB int) error {
	diagnostics.Infof("Creating raw disk image: %s of size %vMB", diskPath, sizeMB)

	_, err := os.Stat(diskPath)
	if err != nil {
		if !os.IsNotExist(err) {
			// un-handle-able error stat-ing the disk file
			return fmt.Errorf("stat: %w", err)
		}
		// disk file does not exist; create it
		file, err := os.OpenFile(diskPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		defer file.Close()

		if err := file.Truncate(util.ConvertMBToBytes(sizeMB)); err != nil {
			return fmt.Errorf("truncate: %w", err)
		}
	}
	return nil
}

// CommonDriver is the common driver base class
type CommonDriver struct {
	// CommandOptions keeps the minikube command line options shared with the
	// minikube packages. This is initialized when loading the driver and should
	// not be persisted.
	// TODO: Consider removing when libmachine API is part of minikube:
	// https://github.com/kubernetes/minikube/issues/21789
	CommandOptions run.CommandOptions `json:"-"`
}

// GetCreateFlags is not implemented yet
func (d *CommonDriver) GetCreateFlags() []mcnflag.Flag {
	return nil
}

// SetConfigFromFlags is not implemented yet
func (d *CommonDriver) SetConfigFromFlags(_ drivers.DriverOptions) error {
	return nil
}

func createRawDiskImage(sshKeyPath, diskPath string, diskSizeMb int) error {
	tarBuf, err := mcnutils.MakeDiskImage(sshKeyPath)
	if err != nil {
		return fmt.Errorf("make disk image: %w", err)
	}

	file, err := os.OpenFile(diskPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer file.Close()
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	if _, err := file.Write(tarBuf.Bytes()); err != nil {
		return fmt.Errorf("write tar: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("closing file %s: %w", diskPath, err)
	}

	if err := os.Truncate(diskPath, util.ConvertMBToBytes(diskSizeMb)); err != nil {
		return fmt.Errorf("truncate: %w", err)
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
	klog.Infof("Making disk image using store path: %s", d.StorePath)
	b2 := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2.CopyIsoToMachineDir(boot2dockerURL, d.MachineName); err != nil {
		return fmt.Errorf("copy iso to machine dir: %w", err)
	}

	keyPath := d.GetSSHKeyPath()
	klog.Infof("Creating ssh key: %s...", keyPath)
	if err := ssh.GenerateSSHKey(keyPath); err != nil {
		return fmt.Errorf("generate ssh key: %w", err)
	}

	diskPath := GetDiskPath(d)
	klog.Infof("Creating raw disk image: %s...", diskPath)
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		if err := createRawDiskImage(publicSSHKeyPath(d), diskPath, diskSize); err != nil {
			return fmt.Errorf("createRawDiskImage(%s): %w", diskPath, err)
		}
		machPath := d.ResolveStorePath(".")
		if err := fixMachinePermissions(machPath); err != nil {
			return fmt.Errorf("fixing permissions on %s: %w", machPath, err)
		}
	}
	return nil
}

func fixMachinePermissions(path string) error {
	klog.Infof("Fixing permissions on %s ...", path)
	if err := os.Chown(path, syscall.Getuid(), syscall.Getegid()); err != nil {
		return fmt.Errorf("chown dir: %w", err)
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}
	for _, f := range files {
		fp := filepath.Join(path, f.Name())
		if err := os.Chown(fp, syscall.Getuid(), syscall.Getegid()); err != nil {
			return fmt.Errorf("chown file: %w", err)
		}
	}
	return nil
}
