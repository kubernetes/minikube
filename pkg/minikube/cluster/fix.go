package cluster

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util/retry"
)

var (
	// The maximum the guest VM clock is allowed to be ahead and behind. This value is intentionally
	// large to allow for inaccurate methodology, but still small enough so that certificates are likely valid.
	maxClockDesyncSeconds = 2.1
)

// fixHost fixes up a previously configured VM so that it is ready to run Kubernetes
func fixHost(api libmachine.API, mc config.MachineConfig) error {
	h, err := api.Load(cfg.Name)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading existing host. Please try running [minikube delete], then run [minikube start] again.")
	}

	s, err := h.Driver.GetState()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting state for host")
	}

	if s == state.Running {
		out.T(out.Running, `Using the running {{.driver_name}} "{{.profile_name}}" VM ...`, out.V{"driver_name": cfg.VMDriver, "profile_name": cfg.Name})
	} else {
		out.T(out.Restarting, `Starting existing {{.driver_name}} VM for "{{.profile_name}}" ...`, out.V{"driver_name": cfg.VMDriver, "profile_name": cfg.Name})
		if err := h.Driver.Start(); err != nil {
			return nil, errors.Wrap(err, "start")
		}
		if err := api.Save(h); err != nil {
			return nil, errors.Wrap(err, "save")
		}
	}

	e := engineOptions(cfg)
	out.T(out.Waiting, "Waiting for the host to be provisioned ...")
	err = fixHost(h, e, cfg)
	if err != nil {
		return nil, err
	}
	return h, nil

	start := time.Now()
	glog.Infof("fixHost: %+v", h.Driver)
	defer func() {
		glog.Infof("fixHost completed within %s", time.Since(start))
	}()

	if len(e.Env) > 0 {
		h.HostOptions.EngineOptions.Env = e.Env
		glog.Infof("Detecting provisioner ...")
		provisioner, err := provision.DetectProvisioner(h.Driver)
		if err != nil {
			return errors.Wrap(err, "detecting provisioner")
		}
		glog.Infof("Provisioning with %s: %+v", provisioner.String(), *h.HostOptions)
		if err := provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
			return errors.Wrap(err, "provision")
		}
	}

	if err := postStartSetup(h, mc); err != nil {
		return errors.Wrap(err, "post-start")
	}

	glog.Infof("Configuring auth for driver %s ...", h.Driver.DriverName())
	if err := h.ConfigureAuth(); err != nil {
		return &retry.RetriableError{Err: errors.Wrap(err, "Error configuring auth on host")}
	}
	return ensureSyncedGuestClock(h)
}

// ensureGuestClockSync ensures that the guest system clock is relatively in-sync
func ensureSyncedGuestClock(h hostRunner) error {
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
