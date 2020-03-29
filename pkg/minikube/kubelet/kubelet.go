/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package kubelet

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/util/retry"
)

// Stop idempotently stops the kubelet
func Stop(cr command.Runner) error {
	return stop(cr, false)
}

// ForceStop idempotently force stops the kubelet
func ForceStop(cr command.Runner) error {
	return stop(cr, true)
}

// stop dempotently stops the kubelet
func stop(cr command.Runner, force bool) error {
	glog.Infof("stopping kubelet ...")
	stp := func() error {
		cmd := exec.Command("sudo", "systemctl", "stop", "kubelet.service")
		if force {
			cmd = exec.Command("sudo", "systemctl", "stop", "-f", "kubelet.service")
		}
		if rr, err := cr.RunCmd(cmd); err != nil {
			return fmt.Errorf("temporary error for %q : %v", rr.Command(), err)
		}
		cmd = exec.Command("sudo", "systemctl", "show", "-p", "SubState", "kubelet")
		rr, err := cr.RunCmd(cmd)
		if err != nil {
			return fmt.Errorf("temporary error: for %q : %v", rr.Command(), err)
		}
		if !strings.Contains(rr.Stdout.String(), "dead") && !strings.Contains(rr.Stdout.String(), "failed") {
			return fmt.Errorf("unexpected kubelet state: %q", rr.Stdout.String())
		}
		return nil
	}

	if err := retry.Expo(stp, 1*time.Second, time.Minute, 2); err != nil {
		return errors.Wrapf(err, "error stopping kubelet")
	}
	return nil
}

// Start starts the kubelet
func Start(cr command.Runner) error {
	glog.Infof("restarting kubelet.service ...")
	c := exec.Command("sudo", "systemctl", "start", "kubelet")
	if _, err := cr.RunCmd(c); err != nil {
		return err
	}
	return nil
}

// Restart restarts the kubelet
func Restart(cr command.Runner) error {
	glog.Infof("restarting kubelet.service ...")
	c := exec.Command("sudo", "systemctl", "restart", "kubelet.service")
	if _, err := cr.RunCmd(c); err != nil {
		return err
	}
	return nil
}

// Check checks on the status of the kubelet
func Check(cr command.Runner) error {
	glog.Infof("checking for running kubelet ...")
	c := exec.Command("sudo", "systemctl", "is-active", "--quiet", "service", "kubelet")
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "check kubelet")
	}
	return nil
}

// Disable disables the Kubelet
func Disable(cr command.Runner) error {
	glog.Infof("disabling kubelet ...")
	c := exec.Command("sudo", "systemctl", "disable", "kubelet")
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "disable")
	}
	return nil
}

// Enable enables the Kubelet
func Enable(cr command.Runner) error {
	glog.Infof("enabling kubelet ...")
	c := exec.Command("sudo", "systemctl", "enable", "kubelet")
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "enable")
	}
	return nil
}
