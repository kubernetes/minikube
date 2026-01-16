//go:build windows

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

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

// KillExisting will kill existing scheduled stops
func KillExisting(profiles []string, options *run.CommandOptions) {
	for _, profile := range profiles {
		if err := killExisting(profile, options); err != nil {
			klog.Errorf("error terminating scheduled stop for profile %s: %v", profile, err)
		}
	}
}

func killExisting(profile string, options *run.CommandOptions) error {
	klog.Infof("trying to kill existing schedule stop for profile %s...", profile)
	api, err := machine.NewAPIClient(options)
	if err != nil {
		return fmt.Errorf("getting api client for profile %s: %w", profile, err)
	}
	h, err := api.Load(profile)
	if err != nil {
		return fmt.Errorf("Error loading existing host. Please try running [minikube delete], then run [minikube start] again.: %w", err)
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return fmt.Errorf("getting command runner: %w", err)
	}
	// restart scheduled stop service in container
	sysManger := sysinit.New(runner)
	if err := sysManger.Stop(constants.ScheduledStopSystemdService); err != nil {
		return fmt.Errorf("stopping schedule-stop service for profile %s: %w", profile, err)
	}
	return nil
}

// to daemonize on windows, we schedule the stop within minikube itself
// starting the minikube-scheduled-stop systemd service kicks off the scheduled stop
func daemonize(profiles []string, duration time.Duration, options *run.CommandOptions) error {
	for _, profile := range profiles {
		if err := startSystemdService(profile, duration, options); err != nil {
			return fmt.Errorf("implementing scheduled stop for %s: %w", profile, err)
		}
	}
	return nil
}

// to start the systemd service, we first have to tell the systemd service how long to sleep for
// before shutting down minikube from within
// we do this by setting the SLEEP environment variable in the environment file to the users
// requested duration
func startSystemdService(profile string, duration time.Duration, options *run.CommandOptions) error {
	// get ssh runner
	klog.Infof("starting systemd service for profile %s...", profile)
	api, err := machine.NewAPIClient(options)
	if err != nil {
		return fmt.Errorf("getting api client for profile %s: %w", profile, err)
	}
	h, err := api.Load(profile)
	if err != nil {
		return fmt.Errorf("Error loading existing host. Please try running [minikube delete], then run [minikube start] again.: %w", err)
	}
	runner, err := machine.CommandRunner(h)
	if err != nil {
		return fmt.Errorf("getting command runner: %w", err)
	}
	if rr, err := runner.RunCmd(exec.Command("sudo", "mkdir", "-p", "/var/lib/minikube/scheduled-stop")); err != nil {
		return fmt.Errorf("creating dirs: %v: %w", rr.Output(), err)
	}
	// update environment file to include duration
	if err := runner.Copy(environmentFile(duration)); err != nil {
		return fmt.Errorf("copying scheduled stop env file: %w", err)
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
