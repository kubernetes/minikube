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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
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

// fixHost fixes up a previously configured VM so that it is ready to run Kubernetes
func fixHost(api libmachine.API, cc *config.ClusterConfig, n *config.Node) (*host.Host, error) {
	start := time.Now()
	klog.Infof("fixHost starting: %s", n.Name)
	defer func() {
		klog.Infof("fixHost completed within %s", time.Since(start))
	}()

	h, err := api.Load(config.MachineName(*cc, *n))
	if err != nil {
		return h, errors.Wrap(err, "Error loading existing host. Please try running [minikube delete], then run [minikube start] again.")
	}
	defer postStartValidations(h, cc.Driver)

	driverName := h.Driver.DriverName()

	// check if need to re-run docker-env
	maybeWarnAboutEvalEnv(driverName, cc.Name)

	h, err = recreateIfNeeded(api, cc, n, h)
	if err != nil {
		return h, err
	}

	// Avoid reprovisioning "native" driver because provision.Detect requires SSH
	if !driver.IsNative(h.Driver.DriverName()) {
		e := engineOptions(*cc)
		h.HostOptions.EngineOptions.Env = e.Env
		err = provisionDockerMachine(h)
		if err != nil {
			return h, errors.Wrap(err, "provision")
		}
	}

	if driver.IsMock(h.DriverName) {
		return h, nil
	}

	if err := postStartSetup(h, *cc); err != nil {
		return h, errors.Wrap(err, "post-start")
	}

	if driver.IsNative(h.Driver.DriverName()) {
		klog.Infof("%s is local, skipping auth/time setup (requires ssh)", driverName)
		return h, nil
	}

	return h, ensureSyncedGuestClock(h, driverName)
}

func recreateIfNeeded(api libmachine.API, cc *config.ClusterConfig, n *config.Node, h *host.Host) (*host.Host, error) {
	machineName := config.MachineName(*cc, *n)
	machineType := driver.MachineType(cc.Driver)
	recreated := false
	s, serr := h.Driver.GetState()

	klog.Infof("recreateIfNeeded on %s: state=%s err=%v", machineName, s, serr)
	if serr != nil || s == state.Stopped || s == state.None {
		// If virtual machine does not exist due to user interrupt cancel(i.e. Ctrl + C), recreate virtual machine
		me, err := machineExists(h.Driver.DriverName(), s, serr)
		if err != nil {
			klog.Infof("machineExists: %t. err=%v", me, err)
		}

		if !me || err == constants.ErrMachineMissing {
			out.Step(style.Shrug, `{{.driver_name}} "{{.cluster}}" {{.machine_type}} is missing, will recreate.`, out.V{"driver_name": cc.Driver, "cluster": machineName, "machine_type": machineType})
			demolish(api, *cc, *n, h)

			klog.Infof("Sleeping 1 second for extra luck!")
			time.Sleep(1 * time.Second)

			h, err = createHost(api, cc, n)
			if err != nil {
				return nil, errors.Wrap(err, "recreate")
			}

			recreated = true
			s, serr = h.Driver.GetState()
		}
	}

	if serr != constants.ErrMachineMissing {
		klog.Warningf("unexpected machine state, will restart: %v", serr)
	}

	if s == state.Running {
		if !recreated {
			out.Step(style.Running, `Updating the running {{.driver_name}} "{{.cluster}}" {{.machine_type}} ...`, out.V{"driver_name": cc.Driver, "cluster": machineName, "machine_type": machineType})
		}
		return h, nil
	}

	if !recreated {
		out.Step(style.Restarting, `Restarting existing {{.driver_name}} {{.machine_type}} for "{{.cluster}}" ...`, out.V{"driver_name": cc.Driver, "cluster": machineName, "machine_type": machineType})
	}
	if err := h.Driver.Start(); err != nil {
		MaybeDisplayAdvice(err, h.DriverName)
		return h, errors.Wrap(err, "driver start")
	}
	if err := saveHost(api, h, cc, n); err != nil {
		return h, err
	}

	return h, nil
}

// maybeWarnAboutEvalEnv wil warn user if they need to re-eval their docker-env, podman-env
// because docker changes the allocated bind ports after restart https://github.com/kubernetes/minikube/issues/6824
func maybeWarnAboutEvalEnv(drver string, name string) {
	if !driver.IsKIC(drver) {
		return
	}
	if os.Getenv(constants.MinikubeActiveDockerdEnv) != "" {
		out.Step(style.Notice, "Noticed you have an activated docker-env on {{.driver_name}} driver in this terminal:", out.V{"driver_name": drver})
		// TODO: refactor docker-env package to generate only eval command per shell. https://github.com/kubernetes/minikube/issues/6887
		out.WarningT(`Please re-eval your docker-env, To ensure your environment variables have updated ports:

	'minikube -p {{.profile_name}} docker-env'

	`, out.V{"profile_name": name})
	}
	if os.Getenv(constants.MinikubeActivePodmanEnv) != "" {
		out.Step(style.Notice, "Noticed you have an activated podman-env on {{.driver_name}} driver in this terminal:", out.V{"driver_name": drver})
		// TODO: refactor podman-env package to generate only eval command per shell. https://github.com/kubernetes/minikube/issues/6887
		out.WarningT(`Please re-eval your podman-env, To ensure your environment variables have updated ports:

	'minikube -p {{.profile_name}} podman-env'

	`, out.V{"profile_name": name})
	}
}

// ensureGuestClockSync ensures that the guest system clock is relatively in-sync
func ensureSyncedGuestClock(h hostRunner, drv string) error {
	if !driver.IsVM(drv) {
		return nil
	}
	d, err := guestClockDelta(h, time.Now())
	if err != nil {
		klog.Warningf("Unable to measure system clock delta: %v", err)
		return nil
	}
	if math.Abs(d.Seconds()) < maxClockDesyncSeconds {
		klog.Infof("guest clock delta is within tolerance: %s", d)
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
	klog.Infof("guest clock: %s", out)
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
	klog.Infof("Guest: %s Remote: %s (delta=%s)", remote, local, d)
	return d, nil
}

// adjustSystemClock adjusts the guest system clock to be nearer to the host system clock
func adjustGuestClock(h hostRunner, t time.Time) error {
	out, err := h.RunSSHCommand(fmt.Sprintf("sudo date -s @%d", t.Unix()))
	klog.Infof("clock set: %s (err=%v)", out, err)
	return err
}

func machineExistsState(s state.State, err error) (bool, error) {
	if s == state.None {
		return false, constants.ErrMachineMissing
	}
	return true, err
}

func machineExistsError(s state.State, err error, drverr error) (bool, error) {
	_ = s // not used
	if err == drverr {
		// if the error matches driver error
		return false, constants.ErrMachineMissing
	}
	return true, err
}

func machineExistsMessage(s state.State, err error, msg string) (bool, error) {
	if s == state.None || (err != nil && err.Error() == msg) {
		// if the error contains the message
		return false, constants.ErrMachineMissing
	}
	return true, err
}

func machineExistsDocker(s state.State, err error) (bool, error) {
	if s == state.Error {
		// if the kic image is not present on the host machine, when user cancel `minikube start`, state.Error will be return
		return false, constants.ErrMachineMissing
	} else if s == state.None {
		// if the kic image is present on the host machine, when user cancel `minikube start`, state.Native will be return
		return false, constants.ErrMachineMissing
	}
	return true, err
}

// machineExists checks if virtual machine does not exist
// if the virtual machine exists, return true
func machineExists(d string, s state.State, err error) (bool, error) {
	if s == state.Running || s == state.Stopped {
		return true, nil
	}
	switch d {
	case driver.HyperKit:
		return machineExistsMessage(s, err, "connection is shut down")
	case driver.HyperV:
		return machineExistsState(s, err)
	case driver.KVM2:
		return machineExistsState(s, err)
	case driver.Native:
		return machineExistsState(s, err)
	case driver.Parallels:
		return machineExistsMessage(s, err, "connection is shut down")
	case driver.VirtualBox:
		return machineExistsError(s, err, virtualbox.ErrMachineNotExist)
	case driver.VMware:
		return machineExistsState(s, err)
	case driver.VMwareFusion:
		return machineExistsState(s, err)
	case driver.Docker:
		return machineExistsDocker(s, err)
	case driver.Mock:
		if s == state.Error {
			return false, constants.ErrMachineMissing
		}
		return true, err
	default:
		return true, err
	}
}
