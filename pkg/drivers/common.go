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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/util"
)

// LeasesPath is the path to dhcpd leases
const LeasesPath = "/var/db/dhcpd_leases"

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
	log.Infof("Creating raw disk image: %s of size %vMB", diskPath, sizeMB)

	_, err := os.Stat(diskPath)
	if err != nil {
		if !os.IsNotExist(err) {
			// un-handle-able error stat-ing the disk file
			return errors.Wrap(err, "stat")
		}
		// disk file does not exist; create it
		file, err := os.OpenFile(diskPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrap(err, "open")
		}
		defer file.Close()

		if err := file.Truncate(util.ConvertMBToBytes(sizeMB)); err != nil {
			return errors.Wrap(err, "truncate")
		}
	}
	return nil
}

// CommonDriver is the common driver base class
type CommonDriver struct{}

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

	if err := os.Truncate(diskPath, util.ConvertMBToBytes(diskSizeMb)); err != nil {
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
	klog.Infof("Making disk image using store path: %s", d.StorePath)
	b2 := mcnutils.NewB2dUtils(d.StorePath)
	if err := b2.CopyIsoToMachineDir(boot2dockerURL, d.MachineName); err != nil {
		return errors.Wrap(err, "copy iso to machine dir")
	}

	keyPath := d.GetSSHKeyPath()
	klog.Infof("Creating ssh key: %s...", keyPath)
	if err := ssh.GenerateSSHKey(keyPath); err != nil {
		return errors.Wrap(err, "generate ssh key")
	}

	diskPath := GetDiskPath(d)
	klog.Infof("Creating raw disk image: %s...", diskPath)
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
	klog.Infof("Fixing permissions on %s ...", path)
	if err := os.Chown(path, syscall.Getuid(), syscall.Getegid()); err != nil {
		return errors.Wrap(err, "chown dir")
	}
	files, err := os.ReadDir(path)
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

// DHCPEntry holds a parsed DNS entry
type DHCPEntry struct {
	Name      string
	IPAddress string
	HWAddress net.HardwareAddr
	ID        string
	Lease     string
}

// GetIPAddressByMACAddress gets the IP address of a MAC address
func GetIPAddressByMACAddress(mac string) (string, error) {
	return getIPAddressFromFile(mac, LeasesPath)
}

func getIPAddressFromFile(mac, path string) (string, error) {
	// Due to https://openradar.appspot.com/FB15382970 we need to parse the MAC
	// address and compare the bytes.
	macAddress, err := parseMAC(mac)
	if err != nil {
		return "", err
	}

	log.Debugf("Searching for %s in %s ...", mac, path)
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	dhcpEntries, err := parseDHCPdLeasesFile(file)
	if err != nil {
		return "", err
	}
	log.Debugf("Found %d entries in %s!", len(dhcpEntries), path)
	for _, dhcpEntry := range dhcpEntries {
		log.Debugf("dhcp entry: %+v", dhcpEntry)
		if dhcpEntry.HWAddress == nil {
			continue
		}
		if bytes.Equal(dhcpEntry.HWAddress, macAddress) {
			log.Debugf("Found match: %s", mac)
			return dhcpEntry.IPAddress, nil
		}
	}
	return "", fmt.Errorf("could not find an IP address for %s", mac)
}

func parseDHCPdLeasesFile(file io.Reader) ([]DHCPEntry, error) {
	var (
		dhcpEntry   *DHCPEntry
		dhcpEntries []DHCPEntry
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "{" {
			dhcpEntry = new(DHCPEntry)
			continue
		} else if line == "}" {
			dhcpEntries = append(dhcpEntries, *dhcpEntry)
			continue
		}

		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid line in dhcp leases file: %s", line)
		}
		key, val := split[0], split[1]
		switch key {
		case "name":
			dhcpEntry.Name = val
		case "ip_address":
			dhcpEntry.IPAddress = val
		case "hw_address":
			// The mac addresses have a '1,' at the start.
			macAddress, err := parseMAC(val[2:])
			if err != nil {
				log.Warnf("unable to parse hw_address in dhcp leases file: %q: %s",
					val[2:], err)
				continue
			}
			dhcpEntry.HWAddress = macAddress
		case "identifier":
			dhcpEntry.ID = val
		case "lease":
			dhcpEntry.Lease = val
		default:
			return dhcpEntries, fmt.Errorf("unable to parse line: %s", line)
		}
	}
	return dhcpEntries, scanner.Err()
}

// parseMAC parse both standard fixeed size MAC address "%02x:..." and the
// variable size MAC address on drawin "%x:...".
func parseMAC(mac string) (net.HardwareAddr, error) {
	hw := make(net.HardwareAddr, 6)
	n, err := fmt.Sscanf(mac, "%x:%x:%x:%x:%x:%x",
		&hw[0], &hw[1], &hw[2], &hw[3], &hw[4], &hw[5])
	if n != len(hw) {
		return nil, fmt.Errorf("invalid MAC address: %q", mac)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to parse MAC address: %q: %s", mac, err)
	}
	return hw, nil
}
