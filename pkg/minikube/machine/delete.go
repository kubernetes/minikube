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

package machine

import (
	"context"
	"os/exec"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

// deleteOrphanedKIC attempts to delete an orphaned docker instance for machines without a config file
// used as last effort clean up not returning errors, won't warn user.
func deleteOrphanedKIC(ociBin string, name string) {
	if ociBin != oci.Podman && ociBin != oci.Docker {
		return
	}

	_, err := oci.ContainerStatus(ociBin, name)
	if err != nil {
		klog.Infof("couldn't inspect container %q before deleting: %v", name, err)
		return
	}
	// allow no more than 5 seconds for deleting the container
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := oci.ShutDown(ociBin, name); err != nil {
		klog.Infof("couldn't shut down %s (might be okay): %v ", name, err)
	}
	cmd := exec.CommandContext(ctx, ociBin, "rm", "-f", "-v", name)
	err = cmd.Run()
	if err == nil {
		klog.Infof("Found stale kic container and successfully cleaned it up!")
	}
}

// DeleteHost deletes the host VM.
// deleteAbandoned will try to delete the machine even if there is no minikube config for it.
func DeleteHost(api libmachine.API, machineName string, deleteAbandoned ...bool) error {
	delAbandoned := true
	if len(deleteAbandoned) > 0 {
		delAbandoned = deleteAbandoned[0]
	}

	hostInfo, err := api.Load(machineName)
	if err != nil && hostInfo == nil && delAbandoned {
		deleteOrphanedKIC(oci.Docker, machineName)
		deleteOrphanedKIC(oci.Podman, machineName)
		// Keep going even if minikube does not know about the host
	}

	// Get the status of the host. Ensure that it exists before proceeding ahead.
	status, err := Status(api, machineName)
	if err != nil {
		// Assume that the host has already been deleted, log and return
		klog.Infof("Unable to get host status for %s, assuming it has already been deleted: %v", machineName, err)
		return nil
	}

	if status == state.None.String() {
		return mcnerror.ErrHostDoesNotExist{Name: machineName}
	}

	// some drivers need manual shut down before delete to avoid getting stuck.
	if driver.NeedsShutdown(hostInfo.Driver.DriverName()) {
		if err := StopHost(api, machineName); err != nil {
			klog.Warningf("stop host: %v", err)
		}
		// Hack: give the Hyper-V VM more time to stop before deletion
		time.Sleep(1 * time.Second)
	}

	out.Step(style.DeletingHost, `Deleting "{{.profile_name}}" in {{.driver_name}} ...`, out.V{"profile_name": machineName, "driver_name": hostInfo.DriverName})
	return deleteHost(api, hostInfo, machineName)
}

// delete removes a host and its local data files
func deleteHost(api libmachine.API, h *host.Host, machineName string) error {
	if err := h.Driver.Remove(); err != nil {
		klog.Warningf("remove failed, will retry: %v", err)
		time.Sleep(1 * time.Second)

		nerr := h.Driver.Remove()
		if nerr != nil {
			return errors.Wrap(nerr, "host remove retry")
		}
	}

	if err := api.Remove(machineName); err != nil {
		return errors.Wrap(err, "api remove")
	}
	return nil
}

// demolish destroys a host by any means necessary - use only if state is inconsistent
func demolish(api libmachine.API, cc config.ClusterConfig, n config.Node, h *host.Host) {
	machineName := config.MachineName(cc, n)
	klog.Infof("DEMOLISHING %s ...", machineName)

	// This will probably fail
	if err := stop(h); err != nil {
		klog.Infof("stophost failed (probably ok): %v", err)
	}

	// For 95% of cases, this should be enough
	if err := DeleteHost(api, machineName); err != nil {
		klog.Warningf("deletehost failed: %v", err)
	}

	if err := deleteHost(api, h, machineName); err != nil {
		klog.Warningf("delete failed (probably ok) %v", err)
	}
}
