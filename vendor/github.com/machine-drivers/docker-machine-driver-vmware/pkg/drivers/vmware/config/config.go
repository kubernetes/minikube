/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

/*
 * Copyright 2017 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package config

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
)

const (
	defaultSSHUser  = "docker"
	defaultSSHPass  = "tcuser"
	defaultDiskSize = 20000
	defaultCPU      = 1
	defaultMemory   = 1024
)

// Config specifies the configuration of driver VMware
type Config struct {
	*drivers.BaseDriver

	Memory         int
	DiskSize       int
	CPU            int
	ISO            string
	Boot2DockerURL string

	SSHPassword    string
	ConfigDriveISO string
	ConfigDriveURL string
	NoShare        bool
}

// NewConfig creates a new Config
func NewConfig(hostname, storePath string) *Config {
	return &Config{
		CPU:         defaultCPU,
		Memory:      defaultMemory,
		DiskSize:    defaultDiskSize,
		SSHPassword: defaultSSHPass,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostname,
			StorePath:   storePath,
		},
	}
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (c *Config) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "VMWARE_BOOT2DOCKER_URL",
			Name:   "vmware-boot2docker-url",
			Usage:  "URL for boot2docker image",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "VMWARE_CONFIGDRIVE_URL",
			Name:   "vmware-configdrive-url",
			Usage:  "URL for cloud-init configdrive",
			Value:  "",
		},
		mcnflag.IntFlag{
			EnvVar: "VMWARE_CPU_COUNT",
			Name:   "vmware-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  defaultCPU,
		},
		mcnflag.IntFlag{
			EnvVar: "VMWARE_MEMORY_SIZE",
			Name:   "vmware-memory-size",
			Usage:  "size of memory for host VM (in MB)",
			Value:  defaultMemory,
		},
		mcnflag.IntFlag{
			EnvVar: "VMWARE_DISK_SIZE",
			Name:   "vmware-disk-size",
			Usage:  "size of disk for host VM (in MB)",
			Value:  defaultDiskSize,
		},
		mcnflag.StringFlag{
			EnvVar: "VMWARE_SSH_USER",
			Name:   "vmware-ssh-user",
			Usage:  "SSH user",
			Value:  defaultSSHUser,
		},
		mcnflag.StringFlag{
			EnvVar: "VMWARE_SSH_PASSWORD",
			Name:   "vmware-ssh-password",
			Usage:  "SSH password",
			Value:  defaultSSHPass,
		},
		mcnflag.BoolFlag{
			EnvVar: "VMWARE_NO_SHARE",
			Name:   "vmware-no-share",
			Usage:  "Disable the mount of your home directory",
		},
	}
}
