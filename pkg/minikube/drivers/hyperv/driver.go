// +build windows

/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package hyperv

import (
	"github.com/docker/machine/drivers/hyperv"
	"github.com/docker/machine/libmachine/drivers"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	registry.Register(registry.DriverDef{
		Name:          constants.DriverHyperv,
		Builtin:       true,
		ConfigCreator: createHypervHost,
		DriverCreator: func() drivers.Driver {
			return hyperv.NewDriver("", "")
		},
	})
}

func createHypervHost(config cfg.MachineConfig) interface{} {
	d := hyperv.NewDriver(cfg.GetMachineName(), localpath.MiniPath())

	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.VSwitch = config.HypervVirtualSwitch
	d.MemSize = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = int(config.DiskSize)
	d.SSHUser = "docker"
	d.DisableDynamicMemory = true // default to disable dynamic memory as minikube is unlikely to work properly with dynamic memory

	return d
}

func ConfigureHostMount(shareName string, hostPath string) error  {
	// Ensure that the current user is administrator because creating a SMB Share requires Administrator privileges.
	_ , err := hyperv.IsWindowsAdministrator()
	if err != nil {
		return err
	}

	// Check if Name of the Share already exists or not.
	if err := hyperv.Cmd("SmbShare\\Get-SmbShare","-Name",shareName); err == nil {
		//log.Debugf("The share with share name %v already exists. Trying to delete it.", shareName)
		if err := hyperv.Cmd("SmbShare\\Remove-SmbShare", "-Name", shareName, "-Force"); err != nil {
			return err
		}
		//log.Debugf("The share with share name %v has been deleted", shareName)
	}

	// Get the current user so that we can assign full access permissions to only that user.
	// TODO - Check if we can use another user.
	user, err := hyperv.GetCurrentWindowsUser()
	if err != nil {
		return err
	}
	// out.T(out.Notice,"Current User -- [{{.user}}]",out.V{"user":user})
	//log.Info("Trying to enable share for CIFS Mounting.")
	if err := hyperv.Cmd("SmbShare\\New-SmbShare", "-Name", shareName, "-Path", hostPath , "-FullAccess", user, "-Temporary"); err != nil {
		return err
	}
	return nil
}