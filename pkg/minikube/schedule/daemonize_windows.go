// +build windows

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

package schedule

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

// KillExisting will kill existing scheduled stops
func KillExisting(profiles []string) {
	for _, profile := range profiles {
		if err := killExisting(profile); err != nil {
			klog.Errorf("error terminating scheduled stop for profile %s: %v", profile, err)
		}
	}
}

func killExisting(profile string) error {
	klog.Infof("trying to kill existing schedule stop for profile %s...", profile)
	api, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrapf(err, "getting api client for profile %s", profile)
	}
	h, err := api.Load(profile)
	if err != nil {
		return errors.Wrap(err, "Error loading existing host. Please try running [minikube delete], then run [minikube start] again.")
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return errors.Wrap(err, "getting command runner")
	}
	// restart scheduled stop service in container
	sysManger := sysinit.New(runner)
	if err := sysManger.Stop(constants.ScheduledStopSystemdService); err != nil {
		return errors.Wrapf(err, "stopping schedule-stop service for profile %s", profile)
	}
	return nil
}

// to daemonize on windows, we schedule the stop within minikube itself
// starting the minikube-scheduled-stop systemd service kicks off the scheduled stop
func daemonize(profiles []string, duration time.Duration) error {
	for _, profile := range profiles {
		if err := startSystemdService(profile, duration); err != nil {
			return errors.Wrapf(err, "implementing scheduled stop for %s", profile)
		}
	}
	return nil
}

// to start the systemd service, we first have to tell the systemd service how long to sleep for
// before shutting down minikube from within
// we do this by settig the SLEEP environment variable in the environment file to the users
// requested duration
func startSystemdService(profile string, duration time.Duration) error {
	// get ssh runner
	klog.Infof("starting systemd service for profile %s...", profile)
	api, err := machine.NewAPIClient()
	if err != nil {
		return errors.Wrapf(err, "getting api client for profile %s", profile)
	}
	h, err := api.Load(profile)
	if err != nil {
		return errors.Wrap(err, "Error loading existing host. Please try running [minikube delete], then run [minikube start] again.")
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return errors.Wrap(err, "getting command runner")
	}
	if rr, err := runner.RunCmd(exec.Command("sudo", "mkdir", "-p", "/var/lib/minikube/scheduled-stop")); err != nil {
		return errors.Wrapf(err, "creating dirs: %v", rr.Output())
	}
	// update environment file to include duration
	if err := runner.Copy(environmentFile(duration)); err != nil {
		return errors.Wrap(err, "copying scheduled stop env file")
	}
	// restart scheduled stop service in container
	sysManager := sysinit.New(runner)
	// enable scheduled stop service
	if err := sysManager.Enable(constants.ScheduledStopSystemdService); err != nil {
		return err
	}
	return sysManager.Restart(constants.ScheduledStopSystemdService)
}

// return the contents of the environment file for minikube-scheduled-stop systemd service
// should be of the format SLEEP=<scheduled stop requested by user in seconds>
func environmentFile(duration time.Duration) assets.CopyableFile {
	contents := []byte(fmt.Sprintf("SLEEP=%v", duration.Seconds()))
	return assets.NewMemoryAssetTarget(contents, constants.ScheduledStopEnvFile, "0644")
}
