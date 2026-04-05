//go:build !windows

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
	"os"
	"strconv"
	"time"

	"github.com/VividCortex/godaemon"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/run"
)

// KillExisting kills existing scheduled stops by looking up the PID
// of the scheduled stop from the PID file saved for the profile and killing the process
func KillExisting(profiles []string, options *run.CommandOptions) {
	for _, profile := range profiles {
		if err := killPIDForProfile(profile); err != nil {
			klog.Warningf("error killng PID for profile %s: %v", profile, err)
		}
		_, cc := mustload.Partial(profile, options)
		cc.ScheduledStop = nil
		if err := config.SaveProfile(profile, cc); err != nil {
			klog.Errorf("error saving profile for profile %s: %v", profile, err)
		}
	}
}

func killPIDForProfile(profile string) error {
	file := localpath.PID(profile)
	f, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return nil
	}
	defer func() {
		if err := os.Remove(file); err != nil {
			klog.Errorf("error deleting %s: %v, you may have to delete in manually", file, err)
		}
	}()
	if err != nil {
		return fmt.Errorf("reading %s: %w", file, err)
	}
	pid, err := strconv.Atoi(string(f))
	if err != nil {
		return fmt.Errorf("converting %s to int: %w", f, err)
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding process: %w", err)
	}
	klog.Infof("killing process %v as it is an old scheduled stop", pid)
	if err := p.Kill(); err != nil {
		return fmt.Errorf("killing %v: %w", pid, err)
	}
	return nil
}

func daemonize(profiles []string, _ time.Duration, _ *run.CommandOptions) error {
	_, _, err := godaemon.MakeDaemon(&godaemon.DaemonAttr{})
	if err != nil {
		return err
	}
	// now that this process has daemonized, it has a new PID
	pid := os.Getpid()
	return savePIDs(pid, profiles)
}

func savePIDs(pid int, profiles []string) error {
	for _, p := range profiles {
		file := localpath.PID(p)
		if err := os.WriteFile(file, []byte(fmt.Sprintf("%v", pid)), 0600); err != nil {
			return err
		}
	}
	return nil
}
