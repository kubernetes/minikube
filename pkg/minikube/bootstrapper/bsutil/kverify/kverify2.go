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

// Package kverify verifies a running kubernetes cluster is healthy
package kverify

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	core "k8s.io/api/core/v1"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/logs"
)

const (
	// minLogCheckTime how long to wait before spamming error logs to console
	minLogCheckTime = 30 * time.Second
	// APIServerWait is the name used in the flags for k8s api server
	APIServerWait = "apiserver"
	// SystemPodsWait is the name used in the flags for pods in the kube system
	SystemPodsWait = "system_pods"
	// DefaultServiceAccountWait is the name used in the flags for default service account
	DefaultServiceAccountWait = "default_sa"
)

// DefaultWaits is map of the the default components to wait for
var DefaultWaits = map[string]bool{APIServerWait: true, SystemPodsWait: true}

// DefaultWaitsKeys is list of all default components to wait for
var DefaultWaitsKeys = []string{APIServerWait, SystemPodsWait}

// NoWaitsCompos is map of componets to wait for if specified 'none' or 'false'
var NoWaitsCompos = map[string]bool{APIServerWait: false, SystemPodsWait: false, DefaultServiceAccountWait: false}

// AllWaitsCompos is map for waiting for all components.
var AllWaitsCompos = map[string]bool{APIServerWait: true, SystemPodsWait: true, DefaultServiceAccountWait: true}

// AllValidWaitsList list of all valid components to wait for
var AllValidWaitsList = []string{APIServerWait, SystemPodsWait, DefaultServiceAccountWait}

// podStatusMsg returns a human-readable pod status, for generating debug status
func podStatusMsg(pod core.Pod) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%q [%s] %s", pod.ObjectMeta.GetName(), pod.ObjectMeta.GetUID(), pod.Status.Phase))
	for i, c := range pod.Status.Conditions {
		if c.Reason != "" {
			if i == 0 {
				sb.WriteString(": ")
			} else {
				sb.WriteString(" / ")
			}
			sb.WriteString(fmt.Sprintf("%s:%s", c.Type, c.Reason))
		}
		if c.Message != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", c.Message))
		}
	}
	return sb.String()
}

// DontWait will return true if the config is no need to wait
func DontWait(wcs map[string]bool) bool {
	return !wcs[APIServerWait] && !wcs[SystemPodsWait] && !wcs[DefaultServiceAccountWait]
}

// announceProblems checks for problems, and slows polling down if any are found
func announceProblems(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner) {
	problems := logs.FindProblems(r, bs, cfg, cr)
	if len(problems) > 0 {
		logs.OutputProblems(problems, 5)
		time.Sleep(kconst.APICallRetryInterval * 15)
	}
}

// KubeletStatus checks the kubelet status
func KubeletStatus(cr command.Runner) (state.State, error) {
	glog.Infof("Checking kubelet status ...")
	rr, err := cr.RunCmd(exec.Command("sudo", "systemctl", "is-active", "kubelet"))
	if err != nil {
		// Do not return now, as we still have parsing to do!
		glog.Warningf("%s returned error: %v", rr.Command(), err)
	}
	s := strings.TrimSpace(rr.Stdout.String())
	glog.Infof("kubelet is-active: %s", s)
	switch s {
	case "active":
		return state.Running, nil
	case "inactive":
		return state.Stopped, nil
	case "activating":
		return state.Starting, nil
	}
	return state.Error, nil
}
