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
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/VividCortex/godaemon"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
)

// Daemonize daemonizes minikube so that scheduled stop happens as expected
func Daemonize(profiles []string, duration time.Duration) error {
	// save current time and expected duration in config
	scheduledStop := &config.ScheduledStopConfig{
		InitiationTime: time.Now().Unix(),
		Duration:       duration,
	}
	for _, p := range profiles {
		// first check if we have a scheduled stop already running that needs to be cancelled
		if err := killExistingScheduledStops(p); err != nil {
			return errors.Wrap(err, "killing existing scheduled stops")
		}
		_, cc := mustload.Partial(p)
		cc.ScheduledStop = scheduledStop
		if err := config.SaveProfile(p, cc); err != nil {
			return errors.Wrap(err, "saving profile")
		}
	}

	_, _, err := godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	if err != nil {
		return errors.Wrap(err, "making daemon")
	}

	// now that this process has daemonized, it has a new PID
	pid := os.Getpid()
	// store this PID in MINIKUBE_HOME/profiles/<profile>/pid
	for _, p := range profiles {
		file := localpath.PID(p)
		if err := ioutil.WriteFile(file, []byte(fmt.Sprintf("%v", pid)), 0644); err != nil {
			return err
		}
	}
	return nil
}

func killExistingScheduledStops(profile string) error {
	file := localpath.PID(profile)
	f, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return nil
	}
	defer os.Remove(file)
	if err != nil {
		return errors.Wrapf(err, "reading %s", file)
	}
	pid, err := strconv.Atoi(string(f))
	if err != nil {
		return errors.Wrapf(err, "converting %v to int", string(f))
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "finding process")
	}
	glog.Infof("killing process %v as it is an old scheduled stop", pid)
	if err := p.Kill(); err != nil {
		return errors.Wrapf(err, "killing %v", pid)
	}
	return nil
}
