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
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
)

// DeleteHost deletes the host VM.
func DeleteHost(api libmachine.API, machineName string) error {
	host, err := api.Load(machineName)
	// Get the status of the host. Ensure that it exists before proceeding ahead.
	status, err := Status(api, machineName)
	if err != nil {
		glog.Infof("Unable to get host status for %s %v, will try to delete anways.", machineName, err)
	}

	if status == state.None.String() {
		glog.Infof("machine %q doesn't exist", machineName)
		return mcnerror.ErrHostDoesNotExist{Name: machineName}
	}

	if driver.NeedsShutdown(host.Driver.DriverName()) {
		if err := StopHost(api, machineName); err != nil {
			glog.Warningf("stop host: %v", err)
		}
		// Hack: give the Hyper-V VM more time to stop before deletion
		time.Sleep(1 * time.Second)
	}

	out.T(out.DeletingHost, `Deleting "{{.profile_name}}" in {{.driver_name}} ...`, out.V{"profile_name": machineName, "driver_name": host.DriverName})
	return delete(api, host, machineName)
}

// delete removes a host and it's local data files
func delete(api libmachine.API, h *host.Host, machineName string) error {
	if err := h.Driver.Remove(); err != nil {
		glog.Warningf("remove failed, will retry: %v", err)
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
	machineName := driver.MachineName(cc, n)
	glog.Infof("DEMOLISHING %s ...", machineName)

	// This will probably fail
	err := stop(h)
	if err != nil {
		glog.Infof("stophost failed (probably ok): %v", err)
	}

	// For 95% of cases, this should be enough
	err = DeleteHost(api, machineName)
	if err != nil {
		glog.Warningf("deletehost failed: %v", err)
	}

	err = delete(api, h, machineName)
	glog.Warningf("delete failed (probably ok) %v", err)
}
