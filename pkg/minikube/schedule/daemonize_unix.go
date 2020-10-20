// +build !windows

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
	"github.com/pkg/errors"
	"k8s.io/klog"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func killExistingScheduledStops(profiles []string) error {
	for _, profile := range profiles {
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
		klog.Infof("killing process %v as it is an old scheduled stop", pid)
		if err := p.Kill(); err != nil {
			return errors.Wrapf(err, "killing %v", pid)
		}
	}
	return nil
}

func daemonize(profiles []string, duration time.Duration) error {
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
		if err := ioutil.WriteFile(file, []byte(fmt.Sprintf("%v", pid)), 0644); err != nil {
			return err
		}
	}
	return nil
}
