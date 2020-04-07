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

// Package kverify verifies a running kubernetes cluster is healthy
package kverify

import (
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

// minLogCheckTime how long to wait before spamming error logs to console
const minLogCheckTime = 60 * time.Second

const (
	// APIServerWaitKey is the name used in the flags for k8s api server
	APIServerWaitKey = "apiserver"
	// SystemPodsWaitKey is the name used in the flags for pods in the kube system
	SystemPodsWaitKey = "system_pods"
	// DefaultSAWaitKey is the name used in the flags for default service account
	DefaultSAWaitKey = "default_sa"
	// AppsRunning is the name used in the flags for waiting for k8s-apps to be running
	AppsRunning = "apps_running"
)

//  vars related to the --wait flag
var (
	// DefaultComponents is map of the the default components to wait for
	DefaultComponents = map[string]bool{APIServerWaitKey: true, SystemPodsWaitKey: true}
	// NoWaitComponents is map of componets to wait for if specified 'none' or 'false'
	NoComponents = map[string]bool{APIServerWaitKey: false, SystemPodsWaitKey: false, DefaultSAWaitKey: false, AppsRunning: false}
	// AllComponents is map for waiting for all components.
	AllComponents = map[string]bool{APIServerWaitKey: true, SystemPodsWaitKey: true, DefaultSAWaitKey: true, AppsRunning: true}
	// DefaultWaitList is list of all default components to wait for. only names to be used for start flags.
	DefaultWaitList = []string{APIServerWaitKey, SystemPodsWaitKey}
	// AllComponentsList list of all valid components keys to wait for. only names to be used used for start flags.
	AllComponentsList = []string{APIServerWaitKey, SystemPodsWaitKey, DefaultSAWaitKey, AppsRunning}
	// AppsRunningList running list are valid k8s-app components to wait for them to be running
	AppsRunningList = []string{
		"kube-dns", // coredns
		"etcd",
		"kube-apiserver",
		"kube-controller-manager",
		"kube-proxy",
		"kube-scheduler",
	}
)

// ShouldWait will return true if the config says need to wait
func ShouldWait(wcs map[string]bool) bool {
	for _, c := range AllComponentsList {
		if wcs[c] {
			return true
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

// KubeletStatus checks the kubelet status
func KubeletStatus(cr command.Runner) (state.State, error) {
	glog.Infof("Checking kubelet status ...")

	active := sysinit.New(cr).Active("kubelet")
	if active {
		return state.Running, nil
	}
	return state.Stopped, nil
	}
	return false
}
