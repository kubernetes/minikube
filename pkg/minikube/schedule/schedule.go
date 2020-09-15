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
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/VividCortex/godaemon"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
)

// Daemonize daemonizes minikube so that scheduled stop happens as expected
func Daemonize(profile string, duration time.Duration) error {
	// save current time and expected duration in config
	api, cc := mustload.Partial(profile)
	defer api.Close()

	cc.ScheduledStop = &config.ScheduledStopConfig{
		InitiationTime: time.Now(),
		Duration:       duration,
	}

	if err := config.SaveProfile(profile, cc); err != nil {
		return errors.Wrap(err, "saving profile")
	}

	_, _, err := godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	if err != nil {
		return err
	}
	// now that this process has daemonized, it has a new PID
	// store this PID in MINIKUBE_HOME/profiles/<profile>/pid
	pid := os.Getpid()
	pidFile := path.Join(localpath.Profile(profile), "pid")
	if err := ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%v", pid)), 0644); err != nil {
		return err
	}

	return nil
}

// Stop schedules a stop for this profile which happens in 'duration' time
func Stop(profile string, duration time.Duration) error {
	sc, err := sleepCommand(profile, duration)
	if err != nil {
		return errors.Wrap(err, "sleep command")
	}
	if err := sc.Start(); err != nil {
		return errors.Wrap(err, "starting command")
	}
	fmt.Println("PID:", sc.Process.Pid)

	godaemon.MakeDaemon(&godaemon.DaemonAttr{})

	// store the time when the command started, the PID, and
	return nil
}

func sleepCommand(profile string, duration time.Duration) (*exec.Cmd, error) {
	// returns the path to the minikube binary being run
	// so that it can be called again later to stop
	currentBinary, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, errors.Wrap(err, "getting path to current binary")
	}
	if runtime.GOOS == "windows" {
		// TODO: priyawadhwa@
		return nil, nil
	}
	cmd := fmt.Sprintf("%s stop -p %s --wait %v", currentBinary, profile, duration.Seconds())
	return exec.Command("sh", "-c", cmd), nil
	// return exec.Command("sh", "-c", "sleep", fmt.Sprintf("%v", duration.Seconds()), "&&", currentBinary, "stop", "-p", profile), nil
}
