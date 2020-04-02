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
	// APIServerWait is the name used in the flags for k8s api server
	APIServerWait = "apiserver"
	// SystemPodsWait is the name used in the flags for pods in the kube system
	SystemPodsWait = "system_pods"
	// DefaultSAWait is the name used in the flags for default service account
	DefaultSAWait = "default_sa"
)

// DefaultWaits is map of the the default components to wait for
var DefaultWaits = map[string]bool{APIServerWait: true, SystemPodsWait: true}

// DefaultWaitsKeys is list of all default components to wait for
var DefaultWaitsKeys = []string{APIServerWait, SystemPodsWait}

// NoWaitsCompo is map of componets to wait for if specified 'none' or 'false'
var NoWaitsCompo = map[string]bool{}

// AllWaitsCompo is map for waiting for all components.
var AllWaitsCompo = map[string]bool{APIServerWait: true, SystemPodsWait: true, DefaultSAWait: true}

// AllValidWaitsList list of all valid components to wait for
var AllValidWaitsList = []string{APIServerWait, SystemPodsWait, DefaultSAWait}

// minLogCheckTime how long to wait before spamming error logs to console
const minLogCheckTime = 30 * time.Second

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

// WaitForSystemPods verifies essential pods for running kurnetes is running
func WaitForSystemPods(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner, client *kubernetes.Clientset, start time.Time, timeout time.Duration) error {
	glog.Info("waiting for kube-system pods to appear ...")
	pStart := time.Now()

	podList := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during pod check")
		}
		if time.Since(start) > minLogCheckTime {
			announceProblems(r, bs, cfg, cr)
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		// Wait for any system pod, as waiting for apiserver may block until etcd
		pods, err := client.CoreV1().Pods("kube-system").List(meta.ListOptions{})
		if err != nil {
			glog.Warningf("pod list returned error: %v", err)
			return false, nil
		}
		glog.Infof("%d kube-system pods found", len(pods.Items))
		for _, pod := range pods.Items {
			glog.Infof(podStatusMsg(pod))
		}

		if len(pods.Items) < 2 {
			return false, nil
		}
		return true, nil
	}
	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, podList); err != nil {
		return fmt.Errorf("apiserver never returned a pod list")
	}
	glog.Infof("duration metric: took %s to wait for pod list to return data ...", time.Since(pStart))
	return nil
}

// WaitForHealthyAPIServer waits for api server status to be running
func WaitForHealthyAPIServer(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner, client *kubernetes.Clientset, start time.Time, hostname string, port int, timeout time.Duration) error {
	glog.Infof("waiting for apiserver healthz status ...")
	hStart := time.Now()

	healthz := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during healthz check")
		}

		if time.Since(start) > minLogCheckTime {
			announceProblems(r, bs, cfg, cr)
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		status, err := apiServerHealthz(hostname, port)
		if err != nil {
			glog.Warningf("status: %v", err)
			return false, nil
		}
		if status != state.Running {
			return false, nil
		}
		return true, nil
	}

	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, healthz); err != nil {
		return fmt.Errorf("apiserver healthz never reported healthy")
	}

	vcheck := func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during version check")
		}
		if err := APIServerVersionMatch(client, cfg.KubernetesConfig.KubernetesVersion); err != nil {
			glog.Warningf("api server version match failed: %v", err)
			return false, nil
		}
		return true, nil
	}

	if err := wait.PollImmediate(kconst.APICallRetryInterval, kconst.DefaultControlPlaneTimeout, vcheck); err != nil {
		return fmt.Errorf("controlPlane never updated to %s", cfg.KubernetesConfig.KubernetesVersion)
	}

	glog.Infof("duration metric: took %s to wait for apiserver health ...", time.Since(hStart))
	return nil
}

// APIServerVersionMatch checks if the server version matches the expected
func APIServerVersionMatch(client *kubernetes.Clientset, expected string) error {
	vi, err := client.ServerVersion()
	if err != nil {
		return errors.Wrap(err, "server version")
	}
	glog.Infof("control plane version: %s", vi)
	if version.CompareKubeAwareVersionStrings(vi.String(), expected) != 0 {
		return fmt.Errorf("controlPane = %q, expected: %q", vi.String(), expected)
	}
	return nil
}

// announceProblems checks for problems, and slows polling down if any are found
func announceProblems(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner) {
	problems := logs.FindProblems(r, bs, cfg, cr)
	if len(problems) > 0 {
		logs.OutputProblems(problems, 5)
		time.Sleep(kconst.APICallRetryInterval * 15)
	}
}

<<<<<<< HEAD
// APIServerStatus returns apiserver status in libmachine style state.State
func APIServerStatus(cr command.Runner, hostname string, port int) (state.State, error) {
	glog.Infof("Checking apiserver status ...")

	pid, err := apiServerPID(cr)
	if err != nil {
		glog.Warningf("stopped: unable to get apiserver pid: %v", err)
		return state.Stopped, nil
	}

	// Get the freezer cgroup entry for this pid
	rr, err := cr.RunCmd(exec.Command("sudo", "egrep", "^[0-9]+:freezer:", fmt.Sprintf("/proc/%d/cgroup", pid)))
	if err != nil {
		glog.Warningf("unable to find freezer cgroup: %v", err)
		return apiServerHealthz(hostname, port)

	}
	freezer := strings.TrimSpace(rr.Stdout.String())
	glog.Infof("apiserver freezer: %q", freezer)
	fparts := strings.Split(freezer, ":")
	if len(fparts) != 3 {
		glog.Warningf("unable to parse freezer - found %d parts: %s", len(fparts), freezer)
		return apiServerHealthz(hostname, port)
	}

	rr, err = cr.RunCmd(exec.Command("sudo", "cat", path.Join("/sys/fs/cgroup/freezer", fparts[2], "freezer.state")))
	if err != nil {
		glog.Errorf("unable to get freezer state: %s", rr.Stderr.String())
		return apiServerHealthz(hostname, port)
	}

	fs := strings.TrimSpace(rr.Stdout.String())
	glog.Infof("freezer state: %q", fs)
	if fs == "FREEZING" || fs == "FROZEN" {
		return state.Paused, nil
	}
	return apiServerHealthz(hostname, port)
}

// apiServerHealthz hits the /healthz endpoint and returns libmachine style state.State
func apiServerHealthz(hostname string, port int) (state.State, error) {
	url := fmt.Sprintf("https://%s/healthz", net.JoinHostPort(hostname, fmt.Sprint(port)))
	glog.Infof("Checking apiserver healthz at %s ...", url)
	// To avoid: x509: certificate signed by unknown authority
	tr := &http.Transport{
		Proxy:           nil, // To avoid connectiv issue if http(s)_proxy is set.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	// Connection refused, usually.
	if err != nil {
		glog.Infof("stopped: %s: %v", url, err)
		return state.Stopped, nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		glog.Errorf("%s returned code %d (unauthorized). Please ensure that your apiserver authorization settings make sense!", url, resp.StatusCode)
		return state.Error, nil
	}
	if resp.StatusCode != http.StatusOK {
		glog.Warningf("%s response: %v %+v", url, err, resp)
		return state.Error, nil
	}
	return state.Running, nil
}

||||||| merged common ancestors
// APIServerStatus returns apiserver status in libmachine style state.State
func APIServerStatus(cr command.Runner, ip net.IP, port int) (state.State, error) {
	glog.Infof("Checking apiserver status ...")

	pid, err := apiServerPID(cr)
	if err != nil {
		glog.Warningf("stopped: unable to get apiserver pid: %v", err)
		return state.Stopped, nil
	}

	// Get the freezer cgroup entry for this pid
	rr, err := cr.RunCmd(exec.Command("sudo", "egrep", "^[0-9]+:freezer:", fmt.Sprintf("/proc/%d/cgroup", pid)))
	if err != nil {
		glog.Warningf("unable to find freezer cgroup: %v", err)
		return apiServerHealthz(ip, port)

	}
	freezer := strings.TrimSpace(rr.Stdout.String())
	glog.Infof("apiserver freezer: %q", freezer)
	fparts := strings.Split(freezer, ":")
	if len(fparts) != 3 {
		glog.Warningf("unable to parse freezer - found %d parts: %s", len(fparts), freezer)
		return apiServerHealthz(ip, port)
	}

	rr, err = cr.RunCmd(exec.Command("sudo", "cat", path.Join("/sys/fs/cgroup/freezer", fparts[2], "freezer.state")))
	if err != nil {
		glog.Errorf("unable to get freezer state: %s", rr.Stderr.String())
		return apiServerHealthz(ip, port)
	}

	fs := strings.TrimSpace(rr.Stdout.String())
	glog.Infof("freezer state: %q", fs)
	if fs == "FREEZING" || fs == "FROZEN" {
		return state.Paused, nil
	}
	return apiServerHealthz(ip, port)
}

// apiServerHealthz hits the /healthz endpoint and returns libmachine style state.State
func apiServerHealthz(ip net.IP, port int) (state.State, error) {
	url := fmt.Sprintf("https://%s/healthz", net.JoinHostPort(ip.String(), fmt.Sprint(port)))
	glog.Infof("Checking apiserver healthz at %s ...", url)
	// To avoid: x509: certificate signed by unknown authority
	tr := &http.Transport{
		Proxy:           nil, // To avoid connectiv issue if http(s)_proxy is set.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	// Connection refused, usually.
	if err != nil {
		glog.Infof("stopped: %s: %v", url, err)
		return state.Stopped, nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		glog.Errorf("%s returned code %d (unauthorized). Please ensure that your apiserver authorization settings make sense!", url, resp.StatusCode)
		return state.Error, nil
	}
	if resp.StatusCode != http.StatusOK {
		glog.Warningf("%s response: %v %+v", url, err, resp)
		return state.Error, nil
	}
	return state.Running, nil
}

=======
>>>>>>> break down kverify package to files
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
