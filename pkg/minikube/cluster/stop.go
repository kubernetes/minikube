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
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util/retry"
)

// StopHost stops the host VM, saving state to disk.
func StopHost(api libmachine.API) error {
	machineName := viper.GetString(config.MachineProfile)
	host, err := api.Load(machineName)
	if err != nil {
		return errors.Wrapf(err, "load")
	}

	out.T(out.Stopping, `Stopping "{{.profile_name}}" in {{.driver_name}} ...`, out.V{"profile_name": machineName, "driver_name": host.DriverName})
	if host.DriverName == driver.HyperV {
		glog.Infof("As there are issues with stopping Hyper-V VMs using API, trying to shut down using SSH")
		if err := trySSHPowerOff(host); err != nil {
			return errors.Wrap(err, "ssh power off")
		}
	}

	if err := host.Stop(); err != nil {
		alreadyInStateError, ok := err.(mcnerror.ErrHostAlreadyInState)
		if ok && alreadyInStateError.State == state.Stopped {
			return nil
		}
		return &retry.RetriableError{Err: errors.Wrapf(err, "Stop: %s", machineName)}
	}
	return nil
}

// trySSHPowerOff runs the poweroff command on the guest VM to speed up deletion
func trySSHPowerOff(h *host.Host) error {
	s, err := h.Driver.GetState()
	if err != nil {
		glog.Warningf("unable to get state: %v", err)
		return err
	}
	if s != state.Running {
		glog.Infof("host is in state %s", s)
		return nil
	}

	out.T(out.Shutdown, `Powering off "{{.profile_name}}" via SSH ...`, out.V{"profile_name": h.Name})
	out, err := h.RunSSHCommand("sudo poweroff")
	// poweroff always results in an error, since the host disconnects.
	glog.Infof("poweroff result: out=%s, err=%v", out, err)
	return nil
}
