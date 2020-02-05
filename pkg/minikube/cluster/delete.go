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

package cluster

import (
	"os/exec"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
)

// deleteOrphanedKIC attempts to delete an orphaned docker instance
func deleteOrphanedKIC(name string) {
	cmd := exec.Command(oci.Docker, "rm", "-f", "-v", name)
	err := cmd.Run()
	if err == nil {
		glog.Infof("Found stale kic container and successfully cleaned it up!")
	}
}

// DeleteHost deletes the host VM.
func DeleteHost(api libmachine.API, machineName string) error {
	host, err := api.Load(machineName)
	if err != nil && host == nil {
		deleteOrphanedKIC(machineName)
		// keep going even if minikube  does not know about the host
	}

	// Get the status of the host. Ensure that it exists before proceeding ahead.
	status, err := GetHostStatus(api, machineName)
	if err != nil {
		// Warn, but proceed
		out.WarningT("Unable to get the status of the {{.name}} cluster.", out.V{"name": machineName})
	}

	if status == state.None.String() {
		return mcnerror.ErrHostDoesNotExist{Name: machineName}
	}

	// This is slow if SSH is not responding, but HyperV hangs otherwise, See issue #2914
	if host.Driver.DriverName() == driver.HyperV {
		if err := trySSHPowerOff(host); err != nil {
			glog.Infof("Unable to power off minikube because the host was not found.")
		}
		out.T(out.DeletingHost, "Successfully powered off Hyper-V. minikube driver -- {{.driver}}", out.V{"driver": host.Driver.DriverName()})
	}

	out.T(out.DeletingHost, `Deleting "{{.profile_name}}" in {{.driver_name}} ...`, out.V{"profile_name": machineName, "driver_name": host.DriverName})
	if err := host.Driver.Remove(); err != nil {
		return errors.Wrap(err, "host remove")
	}
	if err := api.Remove(machineName); err != nil {
		return errors.Wrap(err, "api remove")
	}
	return nil
}
