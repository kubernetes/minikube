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
	"path"
	"time"

	"github.com/VividCortex/godaemon"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
)

// Daemonize daemonizes minikube so that scheduled stop happens as expected
func Daemonize(profiles []string, duration time.Duration) error {
	// save current time and expected duration in config
	scheduledStop := &config.ScheduledStopConfig{
		InitiationTime: time.Now(),
		Duration:       duration,
	}
	for _, p := range profiles {
		api, cc := mustload.Partial(p)
		defer api.Close()

		cc.ScheduledStop = scheduledStop
		if err := config.SaveProfile(p, cc); err != nil {
			return errors.Wrap(err, "saving profile")
		}
	}

	_, _, err := godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	if err != nil {
		return err
	}

	pid := os.Getpid()
	// now that this process has daemonized, it has a new PID
	// store this PID in MINIKUBE_HOME/profiles/<profile>/pid
	for _, p := range profiles {
		pidFile := path.Join(localpath.Profile(p), "pid")
		if err := ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%v", pid)), 0644); err != nil {
			return err
		}
	}
	return nil
}
