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
	"fmt"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/libmachine"
	"k8s.io/minikube/pkg/libmachine/host"
	"k8s.io/minikube/pkg/libmachine/mcnerror"
	"k8s.io/minikube/pkg/libmachine/state"
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
		return fmt.Errorf("load: %w", err)
	}

	out.Step(style.Stopping, `Stopping node "{{.name}}"  ...`, out.V{"name": machineName})
	return stop(h)
}

// stop forcibly stops a host without needing to load
func stop(h *host.Host) error {
	start := time.Now()

	// Windows guests don't have sudo/rsync/poweroff; skip Linux-only SSH operations.
	if driver.IsVM(h.DriverName) && !h.Guest.IsWindows() {
		if err := backup(*h, []string{"/etc/cni", "/etc/kubernetes"}); err != nil {
			klog.Warningf("failed to complete vm config backup (will continue): %v", err)
		}
	}

	if driver.NeedsShutdown(h.DriverName) {
		klog.Infof("GuestOS: %s", h.Guest.Name)
		if err := trySSHPowerOff(h); err != nil {
			return fmt.Errorf("ssh power off: %w", err)
		}
	}

	// For Windows, 'shutdown /s /t 0' already initiated a graceful OS shutdown.
	// Hyper-V\Stop-VM (graceful ACPI) conflicts with an in-progress OS shutdown
	// and returns 0x8007045B. Use Kill (Stop-VM -TurnOff) instead — it succeeds
	// regardless of OS state and is safe since shutdown is already underway.
	stopVM := h.Stop
	if h.Guest.IsWindows() && driver.NeedsShutdown(h.DriverName) {
		stopVM = h.Kill
	}
	if err := stopVM(); err != nil {
		klog.Infof("stop err: %v", err)
		st, ok := err.(mcnerror.ErrHostAlreadyInState)
		if ok && st.State == state.Stopped {
			klog.Infof("host is already stopped")
			return nil
		}
		return &retry.RetriableError{Err: fmt.Errorf("stop: %w", err)}
	}

	klog.Infof("duration metric: took %s to stop", time.Since(start))
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
	// different for kic because RunSSHCommand is not implemented by kic
	if driver.IsKIC(h.DriverName) {
		err := oci.ShutDown(h.DriverName, h.Name)
		klog.Infof("shutdown container: err=%v", err)
	} else if h.Guest.IsWindows() {
		// Windows doesn't have sudo/poweroff; use the Windows shutdown command instead.
		// shutdown /s /t 0 triggers an immediate graceful OS shutdown, so Stop-VM
		// finds the VM already off rather than waiting 30+ seconds for ACPI shutdown.
		rest, err := h.RunSSHCommand("shutdown /s /t 0")
		klog.Infof("windows shutdown result: out=%s, err=%v", rest, err)
	} else {
		rest, err := h.RunSSHCommand("sudo poweroff")
		// poweroff always results in an error, since the host disconnects.
		klog.Infof("poweroff result: out=%s, err=%v", rest, err)
	}
	return nil
}
