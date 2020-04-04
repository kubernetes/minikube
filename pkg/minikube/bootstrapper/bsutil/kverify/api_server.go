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
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	kconst "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
)

// WaitForAPIServerProcess waits for api server to be healthy returns error if it doesn't
func WaitForAPIServerProcess(r cruntime.Manager, bs bootstrapper.Bootstrapper, cfg config.ClusterConfig, cr command.Runner, start time.Time, timeout time.Duration) error {
	glog.Infof("waiting for apiserver process to appear ...")
	err := wait.PollImmediate(time.Millisecond*500, timeout, func() (bool, error) {
		if time.Since(start) > timeout {
			return false, fmt.Errorf("cluster wait timed out during process check")
		}

		if time.Since(start) > minLogCheckTime {
			announceProblems(r, bs, cfg, cr)
			time.Sleep(kconst.APICallRetryInterval * 5)
		}

		if _, ierr := apiServerPID(cr); ierr != nil {
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("apiserver process never appeared")
	}
	glog.Infof("duration metric: took %s to wait for apiserver process to appear ...", time.Since(start))
	return nil
}

// apiServerPID returns our best guess to the apiserver pid
func apiServerPID(cr command.Runner) (int, error) {
	rr, err := cr.RunCmd(exec.Command("sudo", "pgrep", "-xnf", "kube-apiserver.*minikube.*"))
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(rr.Stdout.String())
	return strconv.Atoi(s)
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
