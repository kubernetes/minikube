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
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util/retry"
)

// hostRunner is a minimal host.Host based interface for running commands
type hostRunner interface {
	RunSSHCommand(string) (string, error)
}

const (
	// The maximum the guest VM clock is allowed to be ahead and behind. This value is intentionally
	// large to allow for inaccurate methodology, but still small enough so that certificates are likely valid.
	maxClockDesyncSeconds = 2.1
)

var (
	// ErrorMachineNotExist is returned when virtual machine does not exist due to user interrupt cancel(i.e. Ctrl + C)
	ErrorMachineNotExist = errors.New("machine does not exist")
)

// fixHost fixes up a previously configured VM so that it is ready to run Kubernetes
func fixHost(api libmachine.API, mc config.ClusterConfig) (*host.Host, error) {
	out.T(out.Waiting, "Reconfiguring existing host ...")

	start := time.Now()
	glog.Infof("fixHost starting: %s", mc.Name)
	defer func() {
		glog.Infof("fixHost completed within %s", time.Since(start))
	}()

	h, err := api.Load(mc.Name)
	if err != nil {
		return h, errors.Wrap(err, "Error loading existing host. Please try running [minikube delete], then run [minikube start] again.")
	}

	s, err := h.Driver.GetState()
	if err != nil || s == state.Stopped || s == state.None {
		// If virtual machine does not exist due to user interrupt cancel(i.e. Ctrl + C), recreate virtual machine
		me, err := machineExists(h.Driver.DriverName(), s, err)
		if !me {
			// If the error is that virtual machine does not exist error, handle error(recreate virtual machine)
			if err == ErrorMachineNotExist {
				// remove virtual machine
				if err := h.Driver.Remove(); err != nil {
					// skip returning error since it may be before docker image pulling(so, no host exist)
					if h.Driver.DriverName() != driver.Docker {
						return nil, errors.Wrap(err, "host remove")
					}
				}
				// remove machine config directory
				if err := api.Remove(mc.Name); err != nil {
					return nil, errors.Wrap(err, "api remove")
				}
				// recreate virtual machine
				out.T(out.Meh, "machine '{{.name}}' does not exist. Proceeding ahead with recreating VM.", out.V{"name": mc.Name})
				h, err = createHost(api, mc)
				if err != nil {
					return nil, errors.Wrap(err, "Error recreating VM")
				}
				// return ErrMachineNotExist err to initialize preExists flag
				return h, ErrorMachineNotExist
			}
			// If the error is not that virtual machine does not exist error, return error
			return nil, errors.Wrap(err, "Error getting state for host")
		}
	}

	if s == state.Running {
		out.T(out.Running, `Using the running {{.driver_name}} "{{.profile_name}}" VM ...`, out.V{"driver_name": mc.Driver, "profile_name": mc.Name})
	} else {
		out.T(out.Restarting, `Starting existing {{.driver_name}} VM for "{{.profile_name}}" ...`, out.V{"driver_name": mc.Driver, "profile_name": mc.Name})
		if err := h.Driver.Start(); err != nil {
			return h, errors.Wrap(err, "driver start")
		}
		if err := api.Save(h); err != nil {
			return h, errors.Wrap(err, "save")
		}
	}

	e := engineOptions(mc)
	if len(e.Env) > 0 {
		h.HostOptions.EngineOptions.Env = e.Env
		glog.Infof("Detecting provisioner ...")
		provisioner, err := provision.DetectProvisioner(h.Driver)
		if err != nil {
			return h, errors.Wrap(err, "detecting provisioner")
		}
		if err := provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
			return h, errors.Wrap(err, "provision")
		}
	}

	if driver.IsMock(h.DriverName) {
		return h, nil
	}

	if err := postStartSetup(h, mc); err != nil {
		return h, errors.Wrap(err, "post-start")
	}

	if driver.BareMetal(h.Driver.DriverName()) {
		glog.Infof("%s is local, skipping auth/time setup (requires ssh)", h.Driver.DriverName())
		return h, nil
	}

	glog.Infof("Configuring auth for driver %s ...", h.Driver.DriverName())
	if err := h.ConfigureAuth(); err != nil {
		return h, &retry.RetriableError{Err: errors.Wrap(err, "Error configuring auth on host")}
	}
	return h, ensureSyncedGuestClock(h, mc.Driver)
}

// ensureGuestClockSync ensures that the guest system clock is relatively in-sync
func ensureSyncedGuestClock(h hostRunner, drv string) error {
	if !driver.IsVM(drv) {
		return nil
	}
	d, err := guestClockDelta(h, time.Now())
	if err != nil {
		glog.Warningf("Unable to measure system clock delta: %v", err)
		return nil
	}
	if math.Abs(d.Seconds()) < maxClockDesyncSeconds {
		glog.Infof("guest clock delta is within tolerance: %s", d)
		return nil
	}
	if err := adjustGuestClock(h, time.Now()); err != nil {
		return errors.Wrap(err, "adjusting system clock")
	}
	return nil
}

// guestClockDelta returns the approximate difference between the host and guest system clock
// NOTE: This does not currently take into account ssh latency.
func guestClockDelta(h hostRunner, local time.Time) (time.Duration, error) {
	out, err := h.RunSSHCommand("date +%s.%N")
	if err != nil {
		return 0, errors.Wrap(err, "get clock")
	}
	glog.Infof("guest clock: %s", out)
	ns := strings.Split(strings.TrimSpace(out), ".")
	secs, err := strconv.ParseInt(strings.TrimSpace(ns[0]), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "atoi")
	}
	nsecs, err := strconv.ParseInt(strings.TrimSpace(ns[1]), 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "atoi")
	}
	// NOTE: In a synced state, remote is a few hundred ms ahead of local
	remote := time.Unix(secs, nsecs)
	d := remote.Sub(local)
	glog.Infof("Guest: %s Remote: %s (delta=%s)", remote, local, d)
	return d, nil
}

// adjustSystemClock adjusts the guest system clock to be nearer to the host system clock
func adjustGuestClock(h hostRunner, t time.Time) error {
	out, err := h.RunSSHCommand(fmt.Sprintf("sudo date -s @%d", t.Unix()))
	glog.Infof("clock set: %s (err=%v)", out, err)
	return err
}

// machineExists checks if virtual machine does not exist
// if the virtual machine exists, return true
func machineExists(d string, s state.State, err error) (bool, error) {
	if s == state.Running || s == state.Stopped {
		return true, nil
	}
	switch d {
	case driver.HyperKit:
		if s == state.None || (err != nil && err.Error() == "connection is shut down") {
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.HyperV:
		if s == state.None {
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.KVM2:
		if s == state.None {
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.None:
		if s == state.None {
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.Parallels:
		if err != nil && err.Error() == "machine does not exist" {
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.VirtualBox:
		if err == virtualbox.ErrMachineNotExist {
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.VMware:
		if s == state.None {
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.VMwareFusion:
		if s == state.None {
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.Docker:
		if s == state.Error {
			// if the kic image is not present on the host machine, when user cancel `minikube start`, state.Error will be return
			return false, ErrorMachineNotExist
		} else if s == state.None {
			// if the kic image is present on the host machine, when user cancel `minikube start`, state.None will be return
			return false, ErrorMachineNotExist
		}
		return true, err
	case driver.Mock:
		if s == state.Error {
			return false, ErrorMachineNotExist
		}
		return true, err
	default:
		return true, err
	}
}
