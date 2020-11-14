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
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/retry"
)

// StopHost stops the host VM, saving state to disk.
func StopHost(api libmachine.API, machineName string) error {
	register.Reg.SetStep(register.Stopping)
	klog.Infof("StopHost: %v", machineName)
	h, err := api.Load(machineName)
	if err != nil {
		return errors.Wrapf(err, "load")
	}

	out.Step(style.Stopping, `Stopping node "{{.name}}"  ...`, out.V{"name": machineName})
	return stop(h)
}

// stop forcibly stops a host without needing to load
func stop(h *host.Host) error {
	start := time.Now()
	if driver.NeedsShutdown(h.DriverName) {
		if err := trySSHPowerOff(h); err != nil {
			return errors.Wrap(err, "ssh power off")
		}
	}

	if driver.IsGeneric(h.DriverName) {
		return nil
	}

	if err := h.Stop(); err != nil {
		klog.Infof("stop err: %v", err)
		st, ok := err.(mcnerror.ErrHostAlreadyInState)
		if ok && st.State == state.Stopped {
			klog.Infof("host is already stopped")
			return nil
		}
		return &retry.RetriableError{Err: errors.Wrap(err, "stop")}
	}
	klog.Infof("duration metric: stop complete within %s", time.Since(start))
	return nil
}

// trySSHPowerOff runs the poweroff command on the guest VM to speed up deletion
func trySSHPowerOff(h *host.Host) error {
	s, err := h.Driver.GetState()
	if err != nil {
		klog.Warningf("unable to get state: %v", err)
		return err
	}
	if s != state.Running {
		klog.Infof("host is in state %s", s)
		return nil
	}

	register.Reg.SetStep(register.PowerOff)
	out.Step(style.Shutdown, `Powering off "{{.profile_name}}" via SSH ...`, out.V{"profile_name": h.Name})
	// differnet for kic because RunSSHCommand is not implemented by kic
	if driver.IsKIC(h.DriverName) {
		err := oci.ShutDown(h.DriverName, h.Name)
		klog.Infof("shutdown container: err=%v", err)
	} else {
		out, err := h.RunSSHCommand("sudo poweroff")
		// poweroff always results in an error, since the host disconnects.
		klog.Infof("poweroff result: out=%s, err=%v", out, err)
	}
	return nil
}
