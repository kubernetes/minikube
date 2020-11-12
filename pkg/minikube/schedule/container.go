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
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

func killExisting(profiles []string) error {
	for _, profile := range profiles {
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
	}
	return nil
}

func killWithin(profiles []string, duration time.Duration) error {
	for _, profile := range profiles {
		if err := startSystemdService(profile, duration); err != nil {
			return errors.Wrapf(err, "implementing scheduled stop for %s", profile)
		}
	}
	return nil
}

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
	// update environment file to include duration
	if err := runner.Copy(environmentFile(duration)); err != nil {
		return errors.Wrap(err, "copying scheduled stop env file")
	}
	// restart scheduled stop service in container
	sysManger := sysinit.New(runner)
	return sysManger.Restart(constants.ScheduledStopSystemdService)
}

func environmentFile(duration time.Duration) assets.CopyableFile {
	contents := []byte(fmt.Sprintf("SLEEP=%v", duration.Seconds()))
	return assets.NewMemoryAssetTarget(contents, constants.ScheduledStopEnvFile, "0644")
}
